// FaithConnect - Main JS
document.addEventListener('DOMContentLoaded', function () {
    // Auto-dismiss flash notifications
    const flashMessages = document.querySelectorAll('.flash-notification');
    flashMessages.forEach(function (msg) {
        setTimeout(function () {
            msg.style.opacity = '0';
            msg.style.transform = 'translateY(-20px)';
            setTimeout(function () { msg.remove(); }, 400);
        }, 4000);
    });

    // Table row click navigation
    document.querySelectorAll('tr[data-href]').forEach(function (row) {
        row.style.cursor = 'pointer';
        row.addEventListener('click', function () {
            window.location.href = this.dataset.href;
        });
    });

    // Confirm dangerous actions
    document.querySelectorAll('[data-confirm]').forEach(function (el) {
        el.addEventListener('click', function (e) {
            if (!confirm(this.dataset.confirm)) {
                e.preventDefault();
            }
        });
    });

    // Active nav highlighting
    const path = window.location.pathname;
    document.querySelectorAll('.nav-item[href]').forEach(function (item) {
        const href = item.getAttribute('href');
        if (href && href !== '/' && path.startsWith(href)) {
            item.classList.add('active');
        }
    });
});
