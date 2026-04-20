# SIPROTEC 5 7SJ85 — Справочник CLI-команд

## Системные команды (BusyBox)

| Команда | Описание | Пример |
|---------|----------|--------|
| `ls [path]` | Список файлов в директории | `ls /flash/config/` |
| `cd <dir>` | Смена директории | `cd /flash/comtrade` |
| `pwd` | Текущая директория | `/home/admin` |
| `cat <file>` | Просмотр содержимого файла | `cat /etc/passwd` |
| `df -h` | Использование дисков | rootfs 32M, /flash 256M |
| `ps` | Список процессов | iec61850_srv, protection_engine |
| `top` | Мониторинг процессов/CPU/RAM | CPU 12.3%, RAM 41/128 MB |
| `uptime` | Время работы устройства | 52 days |
| `exit` | Завершение SSH-сессии | — |
| `reboot` | Перезагрузка (запрещено) | Error: Operation not permitted |

## Сетевые команды

| Команда | Описание |
|---------|----------|
| `ifconfig` | Конфигурация сетевых интерфейсов (eth0, eth1, lo) |
| `ping <host>` | Проверка связи (ICMP) |
| `route -n` | Таблица маршрутизации |
| `netstat -tlnp` | Активные порты: 22(SSH), 80(Web), 102(MMS), 161(SNMP), 123(NTP) |

## Инженерные команды IED

| Команда | Описание |
|---------|----------|
| `show_device` | Информация об устройстве (модель, серийник, прошивка, uptime) |
| `show_measurements` | Текущие измерения: I, U, f, P, Q, cosφ |
| `show_settings` | Таблица уставок защит (МТЗ, ТО, ТЗНП, АПВ, ЗНН, ЗМН) |
| `show_protection_status` | Состояние функций защиты (READY/TRIP/BLOCKED) |
| `show_logic_status` | Состояние CFC-логики (цикл, перерасходы) |
| `show_firmware` | Информация о прошивке (активная/резервная, подписи) |
| `show_goose` | Конфигурация GOOSE publisher/subscriber |
| `show_sv` | Конфигурация Sampled Values subscriber |

## Команды защиты и диагностики

| Команда | Описание |
|---------|----------|
| `read_soe` | Последние 15 записей журнала SOE |
| `read_oscilloscope` | Список COMTRADE-записей (осциллограмм) |
| `check_integrity` | Проверка целостности прошивки и конфигурации (SHA-256, CRC32) |
| `iec61850_status` | Статус стека IEC 61850: MMS, GOOSE, SV, Reporting |

## Опасные команды

| Команда | Описание |
|---------|----------|
| `force_output` | Принудительная активация выхода (требует пароль Level-2) |
| `upload_fw` | Загрузка прошивки (отключено в CLI, только через DIGSI 5) |

## Блокируемые команды

| Команда | Ответ |
|---------|-------|
| `sudo *` | `-sh: sudo: not found` |
| `apt/yum/pip *` | `-sh: apt: not found` |
| `rm *` | `Error: Read-only file system` |
| `chmod/chown *` | `Error: Read-only file system` |
| `wget/curl *` | `-sh: wget: not found` |

## Файловая система

```
/
├── bin/            — busybox (ls, cat, df, ps, top, vi, ping)
├── etc/            — hostname, passwd, network.xml, ntp.conf
├── flash/
│   ├── config/     — device_config.cid, substation.scd, relay_settings.cfg
│   ├── comtrade/   — записи осциллограмм (.cfg + .dat)
│   ├── firmware/   — текущая и резервная прошивки
│   └── logic/      — CFC-логика (protection_logic.cfc)
├── var/log/        — soe.log, security.log, system.log
├── proc/           — cpuinfo, meminfo, modules, uptime
└── tmp/
```
