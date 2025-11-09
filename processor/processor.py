import pandas as pd
import sqlite3
import io
import sys
import os

def init_db():
    """Инициализация базы данных"""
    # Используем абсолютный путь к базе данных
    db_path = os.path.join(os.path.dirname(__file__), '../attendance.db')
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS students (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            group_name TEXT NOT NULL,
            student_id TEXT UNIQUE NOT NULL
        )
    ''')
    conn.commit()
    conn.close()
    print(f"Database initialized at: {db_path}")

def process_excel_file(file_content, file_type="students"):
    """
    Обрабатывает Excel файл и сохраняет данные в БД
    
    Args:
        file_content: bytes - содержимое файла
        file_type: str - тип данных (students, topics, supervisors)
    
    Returns:
        dict - результат обработки
    """
    try:
        # Парсим Excel из памяти
        df = pd.read_excel(io.BytesIO(file_content))
        
        print(f"Processing {file_type} file with columns: {list(df.columns)}")
        print(f"Number of rows: {len(df)}")
        
        # Определяем требуемые колонки в зависимости от типа
        if file_type == "students":
            required = ['name', 'group', 'student_id']
            table_name = 'students'
        elif file_type == "topics":
            required = ['topic_name', 'description', 'supervisor']
            table_name = 'topics'
        elif file_type == "supervisors":
            required = ['name', 'department', 'email']
            table_name = 'supervisors'
        else:
            return {"error": f"Unknown file type: {file_type}"}
        
        # Проверяем наличие обязательных колонок
        missing_columns = [col for col in required if col not in df.columns]
        if missing_columns:
            return {"error": f"Missing required columns: {', '.join(missing_columns)}"}
        
        # Подключаемся к БД
        db_path = os.path.join(os.path.dirname(__file__), '../attendance.db')
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        
        # Создаем таблицу если её нет
        if file_type == "students":
            cursor.execute('''
                CREATE TABLE IF NOT EXISTS students (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    name TEXT NOT NULL,
                    group_name TEXT NOT NULL,
                    student_id TEXT UNIQUE NOT NULL
                )
            ''')
        elif file_type == "topics":
            cursor.execute('''
                CREATE TABLE IF NOT EXISTS topics (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    topic_name TEXT NOT NULL,
                    description TEXT,
                    supervisor TEXT NOT NULL
                )
            ''')
        elif file_type == "supervisors":
            cursor.execute('''
                CREATE TABLE IF NOT EXISTS supervisors (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    name TEXT NOT NULL,
                    department TEXT NOT NULL,
                    email TEXT UNIQUE NOT NULL
                )
            ''')
        
        # Вставляем данные
        count = 0
        errors = 0
        
        for index, row in df.iterrows():
            try:
                if file_type == "students":
                    cursor.execute(
                        'INSERT OR IGNORE INTO students (name, group_name, student_id) VALUES (?, ?, ?)',
                        (str(row['name']), str(row['group']), str(row['student_id']))
                    )
                elif file_type == "topics":
                    cursor.execute(
                        'INSERT OR IGNORE INTO topics (topic_name, description, supervisor) VALUES (?, ?, ?)',
                        (str(row['topic_name']), str(row.get('description', '')), str(row['supervisor']))
                    )
                elif file_type == "supervisors":
                    cursor.execute(
                        'INSERT OR IGNORE INTO supervisors (name, department, email) VALUES (?, ?, ?)',
                        (str(row['name']), str(row['department']), str(row['email']))
                    )
                
                if cursor.rowcount > 0:
                    count += 1
                else:
                    errors += 1  # Дубликат или другая ошибка вставки
                    
            except Exception as e:
                print(f"Error inserting row {index}: {e}")
                errors += 1
                continue
        
        conn.commit()
        conn.close()
        
        return {
            "success": True,
            "imported": count,
            "errors": errors,
            "total_rows": len(df),
            "message": f"Successfully imported {count} {file_type} out of {len(df)} rows"
        }
        
    except Exception as e:
        print(f"Error processing file: {e}")
        return {"error": f"Error processing file: {str(e)}"}

def get_stats():
    """Получает статистику из базы данных"""
    try:
        db_path = os.path.join(os.path.dirname(__file__), '../attendance.db')
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        
        stats = {}
        
        # Считаем студентов
        cursor.execute('SELECT COUNT(*) FROM students')
        stats['total_students'] = cursor.fetchone()[0]
        
        # Считаем темы (если таблица существует)
        try:
            cursor.execute('SELECT COUNT(*) FROM topics')
            stats['total_topics'] = cursor.fetchone()[0]
        except:
            stats['total_topics'] = 0
            
        # Считаем руководителей (если таблица существует)
        try:
            cursor.execute('SELECT COUNT(*) FROM supervisors')
            stats['total_supervisors'] = cursor.fetchone()[0]
        except:
            stats['total_supervisors'] = 0
            
        conn.close()
        return stats
        
    except Exception as e:
        print(f"Error getting stats: {e}")
        return {}

if __name__ == "__main__":
    # Тестирование
    init_db()
    if len(sys.argv) > 1:
        with open(sys.argv[1], 'rb') as f:
            content = f.read()
        result = process_excel_file(content, "students")
        print(result)
    else:
        print("Usage: python processor.py <excel_file>")