const http = require('http');

const server = http.createServer((req, res) => {
    res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
    res.end(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>Тест порта 3000</title>
            <style>
                body { font-family: Arial; text-align: center; padding: 50px; background: #000; color: #0f0; }
                h1 { color: #0f0; }
                p { font-size: 20px; }
                .success { color: #0f0; font-weight: bold; }
                .info { color: #ff0; }
            </style>
        </head>
        <body>
            <h1>✅ Порт 3000 РАБОТАЕТ!</h1>
            <p>Если вы видите эту страницу, значит порт 3000 открыт и доступен.</p>
            <p class="success">Подключение с телефона успешно!</p>
            <p class="info">Время: ${new Date().toLocaleString()}</p>
            <p>IP: ${req.socket.remoteAddress}</p>
        </body>
        </html>
    `);
});

server.listen(3000, '0.0.0.0', () => {
    console.log('✅ Сервер запущен на http://0.0.0.0:3000');
    console.log('📱 На телефоне откройте: http://10.184.245.46:3000');
});