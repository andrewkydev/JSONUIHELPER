#JSONUIHELPER

## English:

### This script allows you to view changes in a folder, then it changes the value in manifest.json and archives the folder, then copies the archive to the specified folder

### This script automatically changes the UUID in the manifest.json file

Example config.json

```json
{
  "watchDir": "/path/to/path", //The folder that needs to be watched for changes
  "zipDir": "/path/to/path", //Folder where to copy the finished archive
  "jsonFile": "manifest.json", //Don't touch
  "zipFileName": "name_pack.zip",
  "zip": "true" // If you set it to false, the script simply copies the finished contents of the folder with the modified manifest.json
}
```


## Русский:

### Этот скрипт позволяет просматривать изменения в папке, затем он изменяет значение в manifest.json и архивирует папку, затем копирует архив в указанную папку


### Этот скрипт автоматически меняет UUID в файле manifest.json

Пример config.json

```json
{
"watchDir": "/path/to/path", //Папка, за которой нужно следить на предмет изменений
"zipDir": "/path/to/path", //Папка, куда копировать готовый архив
"jsonFile": "manifest.json", //Не трогать
"zipFileName": "name_pack.zip",
"zip": "true" // Если установить значение false, скрипт просто копирует готовое содержимое папки с измененным manifest.json
}
```