// document.addEventListener('DOMContentLoaded', function() {
//     // Проверяем существование элементов перед добавлением обработчиков
//     const uploadBtn = document.getElementById('uploadBtn');
//     const fileInput = document.getElementById('fileInput');
    
//     // Если кнопка uploadBtn существует, добавляем обработчик
//     if (uploadBtn && fileInput) {
//         uploadBtn.addEventListener('click', async function() {
//             if (fileInput.files.length > 0) {
//                 const originalHTML = this.innerHTML;
//                 this.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Загрузка...';
//                 this.disabled = true;
                
//                 const formData = new FormData();
//                 formData.append('file', fileInput.files[0]);
//                 formData.append('type', 'students');
                
//                 try {
//                     const response = await fetch('/upload', {
//                         method: 'POST',
//                         body: formData
//                     });
                    
//                     const result = await response.json();
                    
//                     if (response.ok) {
//                         this.innerHTML = '<i class="fas fa-check"></i> Успешно!';
//                         this.style.background = '#4CAF50';
                        
//                         if (result.imported) {
//                             console.log(`Загружено ${result.imported} студентов`);
//                         }
//                     } else {
//                         throw new Error(result.error || 'Ошибка загрузки');
//                     }
                    
//                 } catch (error) {
//                     this.innerHTML = '<i class="fas fa-exclamation-triangle"></i> Ошибка';
//                     this.style.background = '#f44336';
//                     alert('Ошибка загрузки: ' + error.message);
//                 }
                
//                 setTimeout(() => {
//                     this.innerHTML = originalHTML;
//                     this.disabled = false;
//                     this.style.background = '';
//                 }, 3000);
//             } else {
//                 alert('Пожалуйста, выберите файл для загрузки');
//             }
//         });
//     }
    
//     // Навигация по меню - проверяем существование элементов
//     const navLinks = document.querySelectorAll('.nav-link');
//     if (navLinks.length > 0) {
//         navLinks.forEach(link => {
//             link.addEventListener('click', function(e) {
//                 e.preventDefault();
//                 document.querySelectorAll('.nav-link').forEach(l => l.classList.remove('active'));
//                 this.classList.add('active');
                
//                 const pageName = this.querySelector('span').textContent;
//                 const pageTitle = document.querySelector('.page-title');
//                 if (pageTitle) {
//                     pageTitle.textContent = pageName;
//                 }
//             });
//         });
//     }
    
//     // Модальное окно - проверяем существование элементов
//     const editModal = document.getElementById('editModal');
//     const closeModal = document.getElementById('closeModal');
//     const cancelBtn = document.getElementById('cancelBtn');
    
//     if (editModal && closeModal && cancelBtn) {
//         const editButtons = document.querySelectorAll('.action-btn .fa-edit');
//         if (editButtons.length > 0) {
//             editButtons.forEach(btn => {
//                 const actionBtn = btn.closest('.action-btn');
//                 if (actionBtn) {
//                     actionBtn.addEventListener('click', function() {
//                         editModal.style.display = 'flex';
//                     });
//                 }
//             });
//         }
        
//         closeModal.addEventListener('click', function() {
//             editModal.style.display = 'none';
//         });
        
//         cancelBtn.addEventListener('click', function() {
//             editModal.style.display = 'none';
//         });
        
//         window.addEventListener('click', function(e) {
//             if (e.target === editModal) {
//                 editModal.style.display = 'none';
//             }
//         });
//     }
// });