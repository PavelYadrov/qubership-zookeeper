#!/bin/bash

set -e

# Названия веток
OLD_BRANCH="old-main"
TEMP_BRANCH="temp-squash"
TARGET_BRANCH="main"

echo "📦 Сохраняю текущее имя коммита HEAD"
HEAD_COMMIT=$(git rev-parse HEAD)

echo "🛑 Создаю резервную ветку $OLD_BRANCH"
git branch -f $OLD_BRANCH $TARGET_BRANCH

echo "🌿 Создаю новую пустую ветку $TEMP_BRANCH без родителей"
git checkout --orphan $TEMP_BRANCH

echo "🧹 Удаляю все отслеживаемые файлы из индекса"
git reset --hard

echo "📥 Копирую файлы из последнего состояния $HEAD_COMMIT"
git checkout $HEAD_COMMIT -- .

echo "✅ Делаю один новый коммит со всеми файлами"
git add .
git commit -m "Initial commit with all project changes"

echo "📌 Переименовываю ветку $TEMP_BRANCH в $TARGET_BRANCH"
git branch -M $TARGET_BRANCH

echo "🚀 Форс-пуш в origin/$TARGET_BRANCH"
git push -f origin $TARGET_BRANCH

echo "🧹 Удаляю резервную ветку $OLD_BRANCH локально"
git branch -D $OLD_BRANCH

echo "🧼 Пытаюсь удалить ветку $OLD_BRANCH на GitHub (если она существует)"
git push origin --delete $OLD_BRANCH || echo "ℹ️ Ветка $OLD_BRANCH не найдена в origin, пропускаю"

echo "✅ Готово! Ветка $TARGET_BRANCH содержит ровно один коммит, история очищена."