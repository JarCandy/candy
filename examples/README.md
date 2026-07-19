# Примеры Caramel

Каждый файл является самостоятельным примером и может быть проверен командой:

```bash
caramel build examples/basics/values.cm
```

| Раздел | Файл | Что показывает |
|---|---|---|
| Основы | [`basics/values.cm`](basics/values.cm) | Строки, числа, логические значения и выражения |
| Основы | [`basics/pointers.cm`](basics/pointers.cm) | Указатели и оператор `&` |
| Коллекции | [`collections/slices.cm`](collections/slices.cm) | Срезы, вложенные срезы и срезы указателей |
| Коллекции | [`collections/maps.cm`](collections/maps.cm) | Map со скалярными и составными значениями |
| Коллекции | [`collections/nested-maps.cm`](collections/nested-maps.cm) | Вложенные map |
| Модели | [`models/profile.cm`](models/profile.cm) | Объявление и создание модели |
| Модели | [`models/relations.cm`](models/relations.cm) | Модели внутри срезов и map |
| Методы | [`methods/user-service.cm`](methods/user-service.cm) | Методы без результата и с несколькими результатами |
| Атрибуты | [`attributes/database.cm`](attributes/database.cm) | Глобальные атрибуты и атрибуты полей |
| Полный пример | [`full/app.cm`](full/app.cm) | Импорты, настройки, модели, коллекции и методы вместе |

Файл [`models/1-model.cm`](models/1-model.cm) содержит расширенный пример синтаксиса, используемый тестами парсера.
