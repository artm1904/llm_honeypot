# IEC 61850: Стандарт для систем автоматизации подстанций

## Общие сведения
IEC 61850 — международный стандарт для коммуникации на подстанциях. Определяет модель данных, протоколы и конфигурацию IED.

## Ключевые протоколы

### MMS (Manufacturing Message Specification)
- Порт: TCP 102
- Назначение: клиент-серверное взаимодействие (чтение/запись данных, управление, отчёты)
- Association: сессия между клиентом (SCADA/DIGSI) и сервером (IED)
- Операции: Read, Write, GetNameList, GetVariableAccessAttributes, Report, Control

### GOOSE (Generic Object Oriented Substation Event)
- Уровень: Ethernet L2 (multicast)
- Назначение: быстрая передача сигналов (релейная защита, блокировка)
- Время: < 4 мс (Performance Class P2/P3)
- Механизм: Publisher/Subscriber
- Поля: AppID, GoCBRef, StNum, SqNum, TAL, Dataset
- VLAN: обязательно (приоритет 4)
- При изменении данных — ретрансляция с уменьшающимся интервалом (2→4→8→...→TAL мс)

### SV (Sampled Values)
- Уровень: Ethernet L2 (multicast)
- Назначение: передача мгновенных значений тока/напряжения от Merging Unit
- SmpRate: 4000 smp/s (80 отсчётов на период при 50 Гц) или 256 smp/s
- Параметры: svID, SmpCnt, SmpSynch, nofASDU

## Модель данных IEC 61850

### Иерархия
```
IED
└── Access Point
    └── Server
        └── Logical Device (LD)
            └── Logical Node (LN)
                └── Data Object (DO)
                    └── Data Attribute (DA)
```

### Основные Logical Nodes
| Группа | LN | Описание |
|--------|-----|----------|
| L | LLN0 | Корневой узел логического устройства |
| L | LPHD | Физическое устройство |
| P | PTOC | Токовая защита (overcurrent) |
| P | PTOV | Защита от перенапряжения |
| P | PTUV | Защита от пониженного напряжения |
| P | PDIS | Дистанционная защита |
| R | RREC | Автоматическое повторное включение (АПВ) |
| X | XCBR | Выключатель (circuit breaker) |
| X | XSWI | Разъединитель (disconnector) |
| M | MMXU | Измерения (токи, напряжения, мощность) |
| M | MMTR | Счётчики энергии |
| C | CSWI | Управление выключателем |
| G | GGIO | Generic I/O |

### Функциональные ограничения (FC)
| FC | Описание |
|----|----------|
| ST | Status (состояние) |
| MX | Measured values (измерения) |
| SP | Setting (уставки) |
| SG | Setting group |
| CF | Configuration |
| DC | Description |
| CO | Control |
| OR | Origin |

## SCL (Substation Configuration Language)
- Формат: XML
- Типы файлов:
  - `.scd` — Substation Configuration Description (полная конфигурация подстанции)
  - `.cid` — Configured IED Description (конфигурация конкретного IED)
  - `.icd` — IED Capability Description (шаблон возможностей IED)
- Секции: Header, Substation, Communication, IED, DataTypeTemplates

## Пример DataSet и Report Control Block
```
DataSet: IED_PTOC_F1/LLN0$PTOC_Trip
  Members:
    IED_PTOC_F1/PTOC1.Op.general   (ST, boolean)
    IED_PTOC_F1/PTOC1.Op.phsA      (ST, boolean)
    IED_PTOC_F1/PTOC1.Str.general   (ST, boolean)

ReportControlBlock: IED_PTOC_F1/LLN0$BR$brcb01
  Dataset: IED_PTOC_F1/LLN0$PTOC_Trip
  TrgOps: dchg, qchg, dupd
  IntgPd: 0 (event-driven)
  OptFlds: seqNum, timeStamp, dataRef, reasonCode
```
