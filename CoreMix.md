# 🚀 Обзор ядер для проксирования и туннелирования

## 🔧 Ядра
<div align="center">

| **Универсальные** | **Специализированные** | **Легковесные** |
|:-:|:-:|:-:|
| `xray` • `v2Fly` | `hysteria2` • `tuic` | `brook` • `overtls` |
| `sing-box` • `mihomo` | `juicity` • `naiveproxy` | |

</div>

---

## 🌐 Поддержка TCP/UDP

| Тип поддержки | Ядра |
|:---|:---|
| **🟢 TCP + UDP** | `xray` `v2Fly` `sing-box` `mihomo` `hysteria2` `tuic` `juicity` `brook` `overtls` |
| **🔴 Только TCP** | `naiveproxy` |
| **⚡ UDP-оптимизированные** | `hysteria2` `tuic` `juicity` |

---

## 📡 Протоколы

### 🔄 Универсальные
```
vmess • vless • trojan • shadowsocks • wireguard
```

### ⚙️ Служебные  
```
dokodemo-door • http • mixed • socks • freedom • dns
```

### 🎯 Специализированные
```
hysteria • hysteria2 • tuic • ssh • tor • shadowtls
```

---

## 🚛 Транспорты

<table>
<tr>
<td width="33%">

### 🔧 Базовые
- `raw` - прямые соединения
- `ws` - WebSocket
- `grpc` - gRPC streaming  
- `http2` - HTTP/2 multiplexing

</td>
<td width="33%">

### ⚡ Продвинутые
- `xhttp` - HTTP long-polling
- `httpupgrade` - HTTP/1.1 upgrade
- `kcp` - надежный UDP
- `quic` - HTTP/3 based

</td>
<td width="33%">

### 🎭 Обфускация
- `reality` - TLS маскировка
- `salamander` - трафик скремблинг
- `brutal` - congestion control

</td>
</tr>
</table>

---

## 🔀 Мультиплексеры

### 🏗️ Встроенные
| Тип | Описание |
|:---|:---|
| **HTTP/2 streams** | Нативное мультиплексирование HTTP/2 |
| **QUIC streams** | QUIC протокол мультиплексирование |
| **gRPC multiplexing** | gRPC потоки |

### 🔌 Внешние
```bash
smux • yamux • h2mux
```

### 🎯 Специализированные
- **Hysteria** → QUIC stream multiplexing
- **TUIC** → native QUIC multiplexing  
- **Sing-box** → multiplex transport layer
- **V2Ray** → mux.cool protocol

---

## 🛡️ Обман DPI (Deep Packet Inspection)

### 🧩 Фрагментация пакетов
```
TCP fragmentation • TLS fragmentation • HTTP header splitting
```

### 🎭 Обфускация заголовков
- **TLS SNI masking** - скрытие Server Name Indication
- **HTTP host header spoofing** - подмена Host заголовка  
- **User-Agent randomization** - случайные браузерные заголовки
- **Fake TLS handshake** - ложные TLS параметры

### ⏱️ Timing атаки
<div align="center">

| Метод | Описание |
|:---:|:---|
| 🎲 **Random delays** | Случайные задержки между пакетами |
| 📊 **Traffic shaping** | Изменение паттернов трафика |
| 💥 **Burst control** | Контроль всплесков трафика |
| 📡 **Jitter injection** | Добавление джиттера |

</div>

### 🎪 Протокольная маскировка
- **Protocol mimicry** - имитация других протоколов
- **Steganography** - скрытие данных в легитимном трафике
- **Decoy traffic** - генерация ложного трафика
- **Traffic mixing** - смешивание с обычным трафиком

---

## 🎭 Маскировка трафика

<table>
<tr>
<td width="50%">

### 🌐 Веб-трафик
- 🔒 **HTTPS** - браузерный трафик
- 🔌 **WebSocket** - WS соединения
- 📡 **HTTP/2** - современный веб
- ⚡ **HTTP/3/QUIC** - новейший стандарт
- 🔧 **gRPC** - API сервисы

### 🎮 Специализированная маскировка
- 🎮 **Gaming UDP** - игровой трафик
- 📺 **Video streaming** - RTMP/HLS потоки
- 📞 **VoIP** - RTP/SRTP трафик
- 🔍 **DNS** - DNS туннелирование
- 📡 **ICMP** - ping пакеты
- 🌊 **BitTorrent** - P2P трафик
- 📁 **FTP** - файловый трафик
- 📧 **SMTP/POP3** - почтовый трафик

</td>
<td width="50%">

### 🚀 Продвинутая маскировка
- 🎯 **Reality** - копирование реальных TLS сайтов
- ☁️ **Domain fronting** - CDN маскировка
- 🔐 **ShadowTLS** - обычные TLS соединения
- 🎭 **Masquerade** - обратный прокси

### 🛡️ Anti-DPI по ядрам

**xray/v2Fly**
```
XTLS Vision • Reality • Fragment • Splice
```

**sing-box**
```
Brutal • Multiplex • ECH • uTLS
```

**mihomo**
```
Dialer • Skip-cert-verify • Interface binding
```

**hysteria2**
```
Salamander • Port hopping • Masquerade
```

</td>
</tr>
</table>

---

## 🔐 TLS Fingerprinting

<div align="center">

### 🌐 Браузеры
| Desktop | Mobile | Other |
|:---:|:---:|:---:|
| Chrome • Firefox | iOS • Android | Edge • QQ |
| Safari | | Random |

### 🛠️ Технологии
```
uTLS • ECH • ALPN • JA3 spoofing • Random fingerprint
```

</div>

---

## 🛡️ Защита

<table>
<tr>
<td width="33%">

### 🔒 Шифрование
- `none` - без шифрования
- `TLS` - стандартное TLS 1.2/1.3
- `Reality` - TLS с маскировкой
- `XTLS` - расширенный TLS

</td>
<td width="33%">

### 🔑 Аутентификация
- `password` - парольная
- `UUID` - уникальные ID
- `certificate pinning` - привязка сертификатов

</td>
<td width="33%">

### 🕵️ Anti-detection
- `fingerprint spoofing` - подмена отпечатков
- `traffic obfuscation` - обфускация трафика

</td>
</tr>
</table>

---

## 🔄 Режимы работы

### 📥 Входящие (Inbound)
```
HTTP/SOCKS proxy • Transparent • Tun • Mixed
```

### 📤 Исходящие (Outbound)  
```
Direct • Proxy chains • Load balancing • Failover
```

### 🔄 Обратный прокси
```
Load balancer • Health checks • Fallback mechanisms
```

---

## ⚡ Оптимизация

<div align="center">

| Категория | Технологии |
|:---:|:---|
| 🚀 **Congestion Control** | BBR • Brutal • Cubic • New Reno |
| 🔀 **Мультиплексирование** | Connection/Stream multiplexing • smux • yamux |
| ⚡ **Специальные** | 0-RTT • Connection migration • Kernel bypass |
| 🛡️ **Anti-DPI** | Fragmentation • Timing randomization • Protocol mimicry |

</div>

---

<div align="center">

### 🎯 **Выбор ядра зависит от ваших потребностей:**
**Универсальность** → `xray` `sing-box`  
**Скорость** → `hysteria2` `tuic`  
**Простота** → `brook` `naiveproxy`  
**Обход блокировок** → `xray Reality` `sing-box`

</div>
