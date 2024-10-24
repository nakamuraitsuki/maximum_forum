#!/bin/bash

# エラーハンドリング
set -e
trap 'echo "エラーが発生しました。全てのプロセスを停止します。"; exit 1' ERR

# backend起動
echo "backend起動中"
cd backend 
go build || { echo "backendのビルドに失敗しました"; exit 1; }
go run main.go &

# バックエンドが起動するまでポートのリッスン状態をチェック（localhost:8080）
echo "バックエンドの起動を待っています..."
while ! nc -z localhost 8080; do   
  sleep 1  # 1秒待機
done

echo "バックエンドが起動しました。次にフロントエンドを起動します。"

# frontend起動
cd ../frontend 
npm start &

wait

