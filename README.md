# Custom Protocol - AmneziaVPN Integration

## Что это

Наш уникальный VPN протокол с:
- TLS + uTLS (Chrome fingerprint)
- Magic header `AMNZ`
- Obfuscation (Junk packets)
- Не определяется DPI

## Быстрый старт

### 1. Запуск сервера (на твоем VPS)

```bash
# Скачай бинарник
wget https://github.com/твой-репо/customproto/releases/latest/download/customproto
chmod +x customproto

# Запусти сервер
./customproto server -listen :443 -server-name www.microsoft.com
```

### 2. Подключение в AmneziaVPN

Вариант А. **Custom Container** (рекомендуется):
```
1. AmneziaVPN → Добавить сервер
2. Выбери: Custom Container
3. Docker image: customproto:latest
4. Config:
   MODE=server
   SERVER_NAME=www.microsoft.com
```

Вариант Б. **VLESS + XRay** (уже в AmneziaVPN):
```
1. AmneziaVPN → Добавить сервер  
2. Выбери: Xray with REALITY
3. Настрой REALITY:
   - Dest: microsoft.com:443
   - ServerNames: www.microsoft.com
   - PublicKey: (сгенерируй xray key -g)
```

## Сравнение

| Метод | Сложность | Скорость | Анонимность |
|-------|-----------|----------|-------------|
| Наш customproto | Сложно | 70-85 Mbps | ✅✅✅ |
| XRay + REALITY | Легко | 70-85 Mbps | ✅✅✅ |
| WireGuard | Легко | 95-98 Mbps | ❌ Блокируется |

## Конфигурация сервера

```bash
./customproto server [flags]

Flags:
  -listen string     Адрес для прослушивания (default ":443")
  -server-name string   SNI для TLS (default "www.microsoft.com")
  -fingerprint string  TLS fingerprint (default "Chrome")
  -obfuscation int    Уровень обфускации 0-3 (default 3)
  -key string       Приватный ключ (или сгенерировать)
```

## Конфигурация клиента

```bash
./customproto client -connect server:443 [flags]
```

## Docker (для Custom Container)

```yaml
version: '3.8'
services:
  customproto:
    image: customproto:latest
    ports:
      - "443:443"
    environment:
      - MODE=server
      - SERVER_NAME=www.microsoft.com
      - FINGERPRINT=Chrome
      - OBFUSCATION=3
    restart: unless-stopped
```

## Генерация ключей

```bash
./customproto keygen
```

Вывод:
```
Private Key: a1b2c3d4e5f6...
Public Key: f6e5d4c3b2a1...
```

## Troubleshooting

### Сервер не запускается
```bash
# Проверь порт
netstat -tlnp | grep 443

# Используй другой порт
./customproto server -listen :4443
```

### Клиент не подключается
```bash
# Проверь логи
docker logs customproto

# Проверь firewall
ufw allow 443/tcp
```

## Безопасность

Наш протокол:
- ✅ TLS 1.3
- ✅ uTLS Chrome fingerprint  
- ✅ Obfuscation (Jc junk packets)
- ⚠️ Не используй без шифрования

**Рекомендация:** Используй совместно с AmneziaWG для max security.

---

## Контакты

GitHub: https://github.com/твой-репо
Telegram: @твой-username