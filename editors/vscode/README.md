# Candy Language Support

Расширение добавляет в Visual Studio Code подсветку синтаксиса Candy для файлов
`.cm` и `.candy`.

Подсвечиваются ключевые слова, комментарии, строки, числа, атрибуты, типы,
квалифицированные пути, вызовы и операторы.

## Разработка

```sh
cd editors/vscode
npm install
npm run check
npm run compile
```

Чтобы запустить расширение локально, откройте каталог `editors/vscode` в VS Code
и нажмите `F5`. В открывшемся Extension Development Host откройте любой файл
`.cm`.

Собрать устанавливаемый файл расширения:

```sh
npm run package
```
