     // Создание анимированного фона с частицами
        function createParticles() {
            const container = document.getElementById('particlesContainer');
            const particleCount = 30;
            
            for (let i = 0; i < particleCount; i++) {
                const particle = document.createElement('div');
                particle.classList.add('particle');
                
                // Случайный размер
                const size = Math.random() * 100 + 50;
                particle.style.width = `${size}px`;
                particle.style.height = `${size}px`;
                
                // Случайная позиция
                const left = Math.random() * 100;
                const top = Math.random() * 100;
                particle.style.left = `${left}%`;
                particle.style.top = `${top}%`;
                
                // Случайная задержка анимации
                const delay = Math.random() * 20;
                particle.style.animationDelay = `${delay}s`;
                
                // Случайная продолжительность анимации
                const duration = Math.random() * 10 + 15;
                particle.style.animationDuration = `${duration}s`;
                
                container.appendChild(particle);
            }
        }

        // Инициализация при загрузке страницы
        document.addEventListener('DOMContentLoaded', function() {
            // Создание анимированного фона
            createParticles();
            
            // Анимация появления элементов
            animateElements();
            
            // Настройка обработчиков событий
            setupEventListeners();
        });

        // Анимация появления элементов
        function animateElements() {
            const statCards = document.querySelectorAll('.stat-card');
            statCards.forEach((card, index) => {
                card.style.opacity = '0';
                card.style.transform = 'translateY(30px)';
                
                setTimeout(() => {
                    card.style.transition = 'all 0.6s ease';
                    card.style.opacity = '1';
                    card.style.transform = 'translateY(0)';
                }, 200 * index);
            });
            
            const tableContainers = document.querySelectorAll('.table-container');
            tableContainers.forEach((container, index) => {
                container.style.opacity = '0';
                container.style.transform = 'translateY(30px)';
                
                setTimeout(() => {
                    container.style.transition = 'all 0.6s ease';
                    container.style.opacity = '1';
                    container.style.transform = 'translateY(0)';
                }, 600 + (200 * index));
            });
        }

        // Настройка обработчиков событий
        function setupEventListeners() {
            // Обработка кнопок быстрых действий
            document.getElementById('viewTopicsBtn').addEventListener('click', function() {
                document.getElementById('topicsTable').scrollIntoView({ behavior: 'smooth' });
            });
            
            document.getElementById('viewStudentsBtn').addEventListener('click', function() {
                document.getElementById('studentsTable').scrollIntoView({ behavior: 'smooth' });
            });
            
            // Обработка модального окна
            document.getElementById('closeAssignModal').addEventListener('click', function() {
                document.getElementById('assignTopicModal').style.display = 'none';
            });
            
            document.getElementById('cancelAssignBtn').addEventListener('click', function() {
                document.getElementById('assignTopicModal').style.display = 'none';
            });
            
            // Закрытие модального окна при клике вне его
            window.addEventListener('click', function(e) {
                if (e.target === document.getElementById('assignTopicModal')) {
                    document.getElementById('assignTopicModal').style.display = 'none';
                }
            });
            
            // Обработка формы назначения темы
            document.getElementById('assignTopicForm').addEventListener('submit', function(e) {
                e.preventDefault();
                assignTopic();
            });
            
            // Обработка поиска
            document.getElementById('studentSearch').addEventListener('input', function() {
                filterTable(this.value, 'studentsTable');
            });
            
            document.getElementById('topicSearch').addEventListener('input', function() {
                filterTable(this.value, 'topicsTable');
            });
            
            // Добавляем интерактивность к карточкам
            const statCards = document.querySelectorAll('.stat-card');
            statCards.forEach(card => {
                card.addEventListener('mouseenter', function() {
                    this.style.transform = 'translateY(-10px) scale(1.02)';
                });
                
                card.addEventListener('mouseleave', function() {
                    this.style.transform = 'translateY(0) scale(1)';
                });
            });
        }

        // Функция фильтрации таблицы
        function filterTable(query, tableId) {
            const table = document.getElementById(tableId);
            const rows = table.getElementsByTagName('tbody')[0].getElementsByTagName('tr');
            
            for (let i = 0; i < rows.length; i++) {
                const cells = rows[i].getElementsByTagName('td');
                let found = false;
                
                for (let j = 0; j < cells.length; j++) {
                    if (cells[j].textContent.toLowerCase().includes(query.toLowerCase())) {
                        found = true;
                        break;
                    }
                }
                
                rows[i].style.display = found ? '' : 'none';
            }
        }

        // Функция назначения темы (интегрируется с бэкендом)
        function assignTopic() {
            const topicId = document.getElementById('assignTopicId').value;
            const studentId = document.getElementById('assignStudent').value;
            const supervisorId = document.getElementById('assignSupervisor').value;
            
            // Отправка данных на сервер
            fetch('/api/assign-topic', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    topicId: topicId,
                    studentId: studentId,
                    supervisorId: supervisorId
                })
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Ошибка назначения темы');
                }
                return response.json();
            })
            .then(data => {
                // Закрытие модального окна
                document.getElementById('assignTopicModal').style.display = 'none';
                
                // Обновление данных
                loadDataFromBackend();
                
                // Показать сообщение об успехе
                showNotification('Тема успешно назначена!', 'success');
            })
            .catch(error => {
                console.error('Ошибка:', error);
                showNotification('Ошибка при назначении темы: ' + error.message, 'error');
            });
        }

        // Функция показа уведомления
        function showNotification(message, type) {
            // Создание элемента уведомления
            const notification = document.createElement('div');
            notification.textContent = message;
            notification.style.position = 'fixed';
            notification.style.top = '20px';
            notification.style.right = '20px';
            notification.style.padding = '15px 20px';
            notification.style.borderRadius = '10px';
            notification.style.color = 'white';
            notification.style.fontWeight = '600';
            notification.style.zIndex = '1000';
            notification.style.boxShadow = '0 5px 15px rgba(0, 0, 0, 0.3)';
            notification.style.transform = 'translateX(100%)';
            notification.style.transition = 'all 0.5s ease';
            
            if (type === 'success') {
                notification.style.background = 'linear-gradient(135deg, #4CAF50, #45a049)';
            } else {
                notification.style.background = 'linear-gradient(135deg, #F44336, #d32f2f)';
            }
            
            document.body.appendChild(notification);
            
            // Анимация появления
            setTimeout(() => {
                notification.style.transform = 'translateX(0)';
            }, 100);
            
            // Автоматическое скрытие через 5 секунд
            setTimeout(() => {
                notification.style.transform = 'translateX(100%)';
                setTimeout(() => {
                    document.body.removeChild(notification);
                }, 500);
            }, 5000);
        }

        // Функции для интеграции с бэкендом

        // Загрузка данных с сервера
        function loadDataFromBackend() {
            // Загрузка статистики
            fetch('/api/stats')
                .then(response => response.json())
                .then(data => updateStatistics(data))
                .catch(error => console.error('Ошибка загрузки статистики:', error));
            
            // Загрузка студентов
            fetch('/api/students')
                .then(response => response.json())
                .then(data => populateStudentsTable(data))
                .catch(error => console.error('Ошибка загрузки студентов:', error));
            
            // Загрузка тем
            fetch('/api/topics')
                .then(response => response.json())
                .then(data => populateTopicsTable(data))
                .catch(error => console.error('Ошибка загрузки тем:', error));
            
            // Загрузка руководителей
            fetch('/api/supervisors')
                .then(response => response.json())
                .then(data => populateSupervisorsDropdown(data))
                .catch(error => console.error('Ошибка загрузки руководителей:', error));
        }

        // Обновление статистики
        function updateStatistics(data) {
            document.getElementById('totalStudents').textContent = data.totalStudents || '0';
            document.getElementById('assignedTopics').textContent = data.assignedTopics || '0';
            document.getElementById('freeTopics').textContent = data.freeTopics || '0';
            document.getElementById('availableSupervisors').textContent = data.availableSupervisors || '0';
        }

        // Заполнение таблицы студентов
        function populateStudentsTable(students) {
            const tbody = document.getElementById('studentsTableBody');
            tbody.innerHTML = '';
            
            students.forEach(student => {
                const row = document.createElement('tr');
                
                // Определение статуса
                let statusClass = 'status-free';
                let statusText = 'Свободен';
                
                if (student.topic && student.supervisor) {
                    statusClass = 'status-assigned';
                    statusText = 'Назначена';
                } else if (student.topic) {
                    statusClass = 'status-taken';
                    statusText = 'Тема занята';
                }
                
                row.innerHTML = `
                    <td>${student.fullName}</td>
                    <td>${student.topic || '-'}</td>
                    <td>${student.supervisor || '-'}</td>
                    <td><span class="status-badge ${statusClass}">${statusText}</span></td>
                    <td class="action-cell">
                        <button class="action-btn" onclick="openAssignModal('${student.id}')" ${student.topic ? 'disabled' : ''}>
                            <i class="fas fa-plus"></i>
                        </button>
                    </td>
                `;
                
                tbody.appendChild(row);
            });
        }

        // Заполнение таблицы тем
        function populateTopicsTable(topics) {
            const tbody = document.getElementById('topicsTableBody');
            tbody.innerHTML = '';
            
            topics.forEach(topic => {
                const row = document.createElement('tr');
                
                // Определение статуса
                let statusClass = topic.status === 'free' ? 'status-free' : 
                                topic.status === 'assigned' ? 'status-assigned' : 'status-taken';
                let statusText = topic.status === 'free' ? 'Свободна' : 
                               topic.status === 'assigned' ? 'Назначена' : 'Занята';
                
                row.innerHTML = `
                    <td>${topic.title}</td>
                    <td>${topic.subject}</td>
                    <td>${topic.workType}</td>
                    <td><span class="status-badge ${statusClass}">${statusText}</span></td>
                    <td class="action-cell">
                        <button class="action-btn" onclick="viewTopicDetails('${topic.id}')">
                            <i class="fas fa-eye"></i>
                        </button>
                        <button class="action-btn" onclick="openAssignModal(null, '${topic.id}')" ${topic.status !== 'free' ? 'disabled' : ''}>
                            <i class="fas fa-plus"></i>
                        </button>
                    </td>
                `;
                
                tbody.appendChild(row);
            });
        }

        // Заполнение выпадающего списка руководителей
        function populateSupervisorsDropdown(supervisors) {
            const select = document.getElementById('assignSupervisor');
            select.innerHTML = '<option value="">-- Выберите руководителя --</option>';
            
            supervisors.forEach(supervisor => {
                const option = document.createElement('option');
                option.value = supervisor.id;
                option.textContent = `${supervisor.fullName} (${supervisor.department})`;
                select.appendChild(option);
            });
        }

        // Открытие модального окна назначения темы
        function openAssignModal(studentId, topicId) {
            document.getElementById('assignTopicId').value = topicId || '';
            
            // Если передали studentId, заполняем выпадающий список студентов
            if (studentId) {
                const select = document.getElementById('assignStudent');
                select.innerHTML = '<option value="">-- Выберите студента --</option>';
                
                // Загружаем список студентов без тем
                fetch('/api/students?withoutTopic=true')
                    .then(response => response.json())
                    .then(students => {
                        students.forEach(student => {
                            const option = document.createElement('option');
                            option.value = student.id;
                            option.textContent = student.fullName;
                            if (student.id === studentId) {
                                option.selected = true;
                            }
                            select.appendChild(option);
                        });
                    })
                    .catch(error => console.error('Ошибка загрузки студентов:', error));
            }
            
            document.getElementById('assignTopicModal').style.display = 'flex';
        }

        // Просмотр деталей темы
        function viewTopicDetails(topicId) {
            // Загрузка деталей темы и отображение в модальном окне
            fetch(`/api/topics/${topicId}`)
                .then(response => response.json())
                .then(topic => {
                    // Создание модального окна с деталями темы
                    // Реализация по необходимости
                    console.log('Детали темы:', topic);
                })
                .catch(error => console.error('Ошибка загрузки деталей темы:', error));
        }

        // Инициализация загрузки данных при загрузке страницы
        window.addEventListener('load', function() {
            loadDataFromBackend();
        });