# 🚀 Обзор ядер для проксирования и туннелирования

## 🔧 Ядра

**Универсальные:** `xray` `v2Fly` `sing-box` `mihomo`  
**Специализированные:** `hysteria2` `tuic` `juicity` `naiveproxy`  
**Легковесные:** `brook` `overtls`

---

## 🌐 Поддержка TCP/UDP

| Поддержка | Ядра |
|---|---|
| **TCP + UDP** | xray, v2Fly, sing-box, mihomo, hysteria2, tuic, juicity, brook, overtls |
| **Только TCP** | naiveproxy |
| **UDP-оптимизированные** | hysteria2, tuic, juicity |

---

## 📡 Протоколы

| Тип | Протоколы |
|---|---|
| **Универсальные** | vmess, vless, trojan, shadowsocks, wireguard |
| **Служебные** | dokodemo-door, http, mixed, socks, freedom, dns |
| **Специализированные** | hysteria, hysteria2, tuic, ssh, tor, shadowtls |

---

## 🚛 Транспорты

| Тип | Транспорты |
|---|---|
| **Базовые** | raw, ws, grpc, http2 |
| **Продвинутые** | xhttp, httpupgrade, kcp, quic |
| **Обфускация** | reality, salamander, brutal |

---

## 🔀 Мультиплексеры

| Тип | Описание |
|---|---|
| **Встроенные** | HTTP/2 streams, QUIC streams, gRPC multiplexing |
| **Внешние** | smux, yamux, h2mux |
| **Специализированные** | Hysteria QUIC, TUIC native, V2Ray mux.cool |

---

## 🛡️ Обман DPI

| Метод | Техники |
|---|---|
| **Фрагментация** | TCP fragmentation, TLS fragmentation, HTTP header splitting |
| **Обфускация заголовков** | SNI masking, Host spoofing, User-Agent randomization |
| **Timing атаки** | Random delays, Traffic shaping, Burst control, Jitter |
| **Протокольная маскировка** | Protocol mimicry, Steganography, Decoy traffic |

---

## 🎭 Маскировка трафика

| Категория | Типы трафика |
|---|---|
| **Веб-трафик** | HTTPS, WebSocket, HTTP/2, HTTP/3/QUIC, gRPC |
| **Специализированная** | Gaming UDP, Video streaming, VoIP, DNS, ICMP, BitTorrent, FTP, SMTP |
| **Продвинутая** | Reality, Domain fronting, ShadowTLS, Masquerade |

---

## 🛡️ Anti-DPI по ядрам

| Ядро | Техники |
|---|---|
| **xray/v2Fly** | XTLS Vision, Reality, Fragment, Splice |
| **sing-box** | Brutal, Multiplex, ECH, uTLS |
| **mihomo** | Dialer, Skip-cert-verify, Interface binding |
| **hysteria2** | Salamander, Port hopping, Masquerade |

---

## 🔐 TLS Fingerprinting

| Тип | Варианты |
|---|---|
| **Браузеры** | Chrome, Firefox, Safari, iOS, Android, Edge, QQ |
| **Технологии** | uTLS, ECH, ALPN, JA3 spoofing, Random fingerprint |

---

## 🛡️ Защита

| Тип | Варианты |
|---|---|
| **Шифрование** | none, TLS, Reality, XTLS |
| **Аутентификация** | password, UUID, certificate pinning |
| **Anti-detection** | fingerprint spoofing, traffic obfuscation |

---

## 🔄 Режимы работы

| Направление | Режимы |
|---|---|
| **Входящие** | HTTP/SOCKS proxy, Transparent, Tun, Mixed |
| **Исходящие** | Direct, Proxy chains, Load balancing, Failover |
| **Обратный прокси** | Load balancer, Health checks, Fallback |

---

## ⚡ Оптимизация

| Категория | Технологии |
|---|---|
| **Congestion Control** | BBR, Brutal, Cubic, New Reno |
| **Мультиплексирование** | Connection/Stream multiplexing, smux, yamux |
| **Специальные** | 0-RTT, Connection migration, Kernel bypass |
| **Anti-DPI** | Fragmentation, Timing randomization, Protocol mimicry |

---

## 🎯 Рекомендации по выбору

| Задача | Рекомендуемые ядра |
|---|---|
| **Универсальность** | xray, sing-box |
| **Высокая скорость** | hysteria2, tuic |
| **Простота настройки** | brook, naiveproxy |
| **Обход блокировок** | xray Reality, sing-box |
