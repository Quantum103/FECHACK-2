function createSnow() {
    const body = document.querySelector('body');
    const snowCount = 50;
    
    for (let i = 0; i < snowCount; i++) {
        const snow = document.createElement('div');
        snow.classList.add('snow');
        
        const size = Math.random() * 5 + 2;
        snow.style.width = size + 'px';
        snow.style.height = size + 'px';
        snow.style.left = Math.random() * 100 + 'vw';
        snow.style.opacity = Math.random() * 0.5 + 0.3;
        snow.style.animationDuration = Math.random() * 10 + 5 + 's';
        snow.style.animationDelay = Math.random() * 5 + 's';
        
        body.appendChild(snow);
    }
}

// Запускаем снегопад
createSnow();