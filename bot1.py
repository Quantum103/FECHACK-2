import sqlite3
import logging
from telegram import Update, ReplyKeyboardMarkup, KeyboardButton
from telegram.ext import Application, CommandHandler, MessageHandler, filters, ContextTypes, ConversationHandler

BOT_TOKEN = "8299175701:AAFgsWHKCEQHAefrwq3dvXOPRwmG81ymZko"
DB_NAME = "attendance.db"

GROUP, FIO = range(2)

logging.basicConfig(
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    level=logging.INFO
)

def init_database():
    """Проверяем структуру базы данных"""
    try:
        conn = sqlite3.connect(DB_NAME)
        cursor = conn.cursor()
        
        # Проверяем существование таблицы users
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='users'")
        if not cursor.fetchone():
            logging.error("❌ Таблица 'users' не найдена в базе данных!")
            return False
            
        # Проверяем структуру таблицы
        cursor.execute("PRAGMA table_info(users)")
        columns = cursor.fetchall()
        logging.info("📊 Структура таблицы users:")
        for column in columns:
            logging.info(f"  - {column[1]} ({column[2]})")
        
        # Проверяем наличие данных
        cursor.execute("SELECT COUNT(*) FROM users")
        count = cursor.fetchone()[0]
        logging.info(f"📊 Количество пользователей в БД: {count}")
        
        conn.close()
        return True
        
    except sqlite3.Error as e:
        logging.error(f"❌ Ошибка подключения к БД: {e}")
        return False

def get_main_keyboard():
    keyboard = [
        [KeyboardButton("📚 Моя тема")],
        [KeyboardButton("🔄 Обновить")]
    ]
    return ReplyKeyboardMarkup(keyboard, resize_keyboard=True)

async def start(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Начало работы с ботом - сразу запрашиваем группу"""
    user_id = update.message.from_user.id
    username = update.message.from_user.username or "Неизвестный"
    
    logging.info(f"👤 Пользователь {username} (ID: {user_id}) запустил бота")
    
    await update.message.reply_text(
        "👋 Добро пожаловать в бот для получения темы курсовой работы!\n\n"
        "🏫 Введите вашу учебную группу (например: ИС-21, ПИ-31):"
    )
    return GROUP

async def get_group(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Получаем группу от пользователя"""
    group = update.message.text.strip()
    context.user_data['group'] = group
    
    await update.message.reply_text(
        f"🏫 Группа: {group}\n\n"
        f"📝 Теперь введите ваше ФИО (например: Александр Игнат Валерьевич):"
    )
    return FIO

async def get_fio(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Получаем ФИО и ищем тему в базе данных"""
    fio = update.message.text.strip()
    user_id = update.message.from_user.id
    group = context.user_data.get('group', 'Не указана')
    
    logging.info(f"🔍 Поиск пользователя: {fio}, группа: {group}")
    
    conn = sqlite3.connect(DB_NAME)
    cursor = conn.cursor()
    
    try:
        # Ищем пользователя по ФИО в базе данных
        # Пробуем разные варианты названий полей
        queries = [
            "SELECT topic FROM users WHERE name = ?",
            "SELECT topic FROM users WHERE Name = ?", 
            "SELECT Topic FROM users WHERE name = ?",
            "SELECT topic FROM users WHERE full_name = ?",
            "SELECT course_topic FROM users WHERE name = ?"
        ]
        
        topic = None
        used_query = ""
        
        for query in queries:
            try:
                cursor.execute(query, (fio,))
                result = cursor.fetchone()
                if result:
                    topic = result[0]
                    used_query = query
                    break
            except sqlite3.Error:
                continue
        
        if topic:
            # Сохраняем telegram_id для этого пользователя
            try:
                update_queries = [
                    "UPDATE users SET telegram_id = ? WHERE name = ?",
                    "UPDATE users SET telegram_id = ? WHERE Name = ?",
                ]
                
                for update_query in update_queries:
                    try:
                        cursor.execute(update_query, (user_id, fio))
                        conn.commit()
                        break
                    except sqlite3.Error:
                        continue
                        
            except sqlite3.Error as e:
                logging.warning(f"⚠️ Не удалось обновить telegram_id: {e}")
            
            await update.message.reply_text(
                f"✅ Найдена ваша тема курсовой работы!\n\n"
                f"📝 ФИО: {fio}\n"
                f"🏫 Группа: {group}\n"
                f"📚 Тема: {topic}",
                reply_markup=get_main_keyboard()
            )
            logging.info(f"✅ Найдена тема для {fio}: {topic}")
        else:
            await update.message.reply_text(
                f"❌ Не удалось найти тему курсовой работы.\n\n"
                f"📝 ФИО: {fio}\n"
                f"🏫 Группа: {group}\n\n"
                f"Возможные причины:\n"
                f"• Проверьте правильность введенного ФИО\n"
                f"• Обратитесь к преподавателю",
                reply_markup=get_main_keyboard()
            )
            logging.warning(f"❌ Тема не найдена для {fio}")
            
    except sqlite3.Error as e:
        logging.error(f"❌ Ошибка БД при поиске темы: {e}")
        await update.message.reply_text(
            "❌ Произошла ошибка при поиске в базе данных. Попробуйте позже."
        )
    finally:
        conn.close()
    
    return ConversationHandler.END

async def show_my_topic(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Показывает тему курсовой работы по команде /my_topic или кнопке"""
    user_id = update.message.from_user.id
    
    conn = sqlite3.connect(DB_NAME)
    cursor = conn.cursor()
    
    try:
        # Ищем пользователя по telegram_id
        cursor.execute("SELECT name, topic FROM users WHERE telegram_id = ?", (user_id,))
        user_data = cursor.fetchone()
        
        if user_data:
            name, topic = user_data
            if topic:
                await update.message.reply_text(
                    f"📚 Ваша тема курсовой работы:\n\n"
                    f"📝 {topic}"
                )
            else:
                await update.message.reply_text(
                    "❌ У вас еще нет назначенной темы курсовой работы."
                )
        else:
            await update.message.reply_text(
                "❌ Сначала зарегистрируйтесь через /start"
            )
            
    except sqlite3.Error as e:
        logging.error(f"❌ Ошибка БД: {e}")
        await update.message.reply_text("❌ Ошибка базы данных")
    finally:
        conn.close()

async def handle_buttons(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Обработка нажатий кнопок"""
    text = update.message.text
    
    if text == "📚 Моя тема":
        await show_my_topic(update, context)
    elif text == "🔄 Обновить":
        await update.message.reply_text("✅ Данные обновлены!")
        await show_my_topic(update, context)

async def cancel(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Отмена регистрации"""
    await update.message.reply_text(
        "Регистрация отменена.\n\nИспользуйте /start для начала работы"
    )
    return ConversationHandler.END

async def debug_info(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Команда для отладки - показывает информацию о базе данных"""
    conn = sqlite3.connect(DB_NAME)
    cursor = conn.cursor()
    
    try:
        # Показываем структуру таблицы users
        cursor.execute("PRAGMA table_info(users)")
        columns = cursor.fetchall()
        
        message = "📊 Структура таблицы users:\n"
        for col in columns:
            message += f"  {col[1]} ({col[2]})\n"
        
        # Показываем первые 5 записей
        cursor.execute("SELECT name, topic FROM users LIMIT 5")
        users = cursor.fetchall()
        
        message += "\n📝 Примеры записей:\n"
        for user in users:
            message += f"  {user[0]} - {user[1]}\n"
        
        await update.message.reply_text(message)
        
    except sqlite3.Error as e:
        await update.message.reply_text(f"❌ Ошибка: {e}")
    finally:
        conn.close()

def main():
    logging.info("🚀 Запуск бота...")
    
    if not init_database():
        logging.error("❌ Не удалось подключиться к базе данных!")
        return
    
    application = Application.builder().token(BOT_TOKEN).build()
    
    # Обработчик диалога регистрации
    conv_handler = ConversationHandler(
        entry_points=[CommandHandler('start', start)],
        states={
            GROUP: [MessageHandler(filters.TEXT & ~filters.COMMAND, get_group)],
            FIO: [MessageHandler(filters.TEXT & ~filters.COMMAND, get_fio)],
        },
        fallbacks=[CommandHandler('cancel', cancel)]
    )
    
    application.add_handler(conv_handler)
    application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_buttons))
    application.add_handler(CommandHandler("my_topic", show_my_topic))
    application.add_handler(CommandHandler("debug", debug_info))  # Для отладки
    
    print("✅ Бот запущен")
    application.run_polling()

if __name__ == '__main__':
    main()