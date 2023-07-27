release:
	fyne package --os darwin --id com.kristas.kuui --icon icon.png --release
release-android:
	fyne package --os android --id com.kristas.kuui --icon icon.png

build:
	go build -ldflags="-w -s"

install:
	go install github.com/kristax/kuui