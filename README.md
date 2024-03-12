## Описание системы
```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml

LAYOUT_WITH_LEGEND()
LAYOUT_LANDSCAPE()

title
  <b>GophKeeperArch v2024.02.17</b>
  <i>Управление паролями GophKeeper</i>
end title

Person(user, "Пользователь")
System(goph_keeper, "Менеджер паролей")

Rel(user, goph_keeper, "Создание\получение\удаление паролей", "Text-User Interface")

@enduml
```

## Описание контейнеров
```plantuml
@startuml
!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Container.puml

LAYOUT_WITH_LEGEND()
LAYOUT_TOP_DOWN()

title
  <b>GophKeeperContainers v2024.02.17</b>
  <i>Описание контейнеров GophKeeper</i>
end title

Person(user, "Пользователь")
System_Boundary(password_managing_desktop, "Клиент системы хранения паролей") {
  Container(goph_keeper_cli, "Менеджер паролей")
  Rel(user, goph_keeper_cli, "Создание\получение\удаление паролей", "Text-User Interface")
}

System_Boundary(passmword_managing_server, "Сервер системы хранения паролей") {
  Container(goph_keeper_server, "Сервер хранения паролей")
  ContainerDb(db, "База данных", "PostgreSQL", "Хранит пользователей, сессии, пароли и т.д.", "")
  Rel_D(goph_keeper_cli, goph_keeper_server, "Запросы к серверу", "HTTP")
  Rel_D(goph_keeper_server, db, "Чтение\Запись")
}

@enduml
```

## Описание компонентов
![schema](https://www.plantuml.com/plantuml/png/PP0nJmGX48LxViLaQit5bi9SQwM9PttXs74NbsEGcLqBS_zTs9FaeXKXlEzxB-n5NT7b78rnNhd08bICvnZ9Q-0aW2FdwJWJPIf77mE24nX8PkNyB_ZHWrBMBLXjwrzXPj6na7p6g-jaMYdSFtOjM3YyFG4bfG8w4KGUGmAN1iXEv8lBO7gqKjUE2hqylnuixDQ7NQ4nIAE_iRcJEnDQGBm3x0QqY1VpoxRkKpNuYVmC7TaQGwethOVozY2TkTrRUvHdIbl9vNPSGop8uZs9l2yH-ZHZzfI6-lC_)


### Регистрация пользователя
```plantuml
@startuml

actor user
collections "goph_keeper_cli" as desktop
collections "gopj_keeper_srv" as server
collections "postgresql" as db
user -> desktop : TUI enter the login and password
desktop -> server : POST: /v1/register {"username": "", "password":""}
server -> db : check username and insert into users
server -> desktop : {"token": ""}
desktop -> user: successful register

@enduml
```

### Аутентификация пользователя
```plantuml
@startuml

actor user
collections "goph_keeper_cli" as desktop
collections "gopj_keeper_srv" as server
collections "postgresql" as db
user -> desktop : TUI enter the login and password
desktop -> server : POST: /v1/login {"username": "", "password":""}
server -> db : check username and create token
server -> desktop : {"token": ""}
desktop -> user: successful login

@enduml
```

### Создание записи с паролем
```plantuml
@startuml

actor user
collections "goph_keeper_cli" as desktop
collections "gopj_keeper_srv" as server
collections "postgresql" as db
user -> desktop : TUI enter the login and password
desktop -> server : POST: /v1/password {"name": "", "password":"", ...}
server -> db : validate password model and insert into password
server -> desktop : 201
desktop -> user: successful create new password

@enduml
```