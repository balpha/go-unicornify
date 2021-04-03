rm -rf release

mkdir -p release/linux
mkdir -p release/windows
mkdir -p release/mac

GOOS=linux GOARCH=amd64 go build -o release/linux/unicornify
GOOS=windows GOARCH=amd64 go build -o release/windows/unicornify.exe
GOOS=darwin GOARCH=amd64 go build -o release/mac/unicornify

zip -j release/unicornify-linux.zip release/linux/*
zip -j release/unicornify-windows.zip release/windows/*
zip -j release/unicornify-mac.zip release/mac/*

