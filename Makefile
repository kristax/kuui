release:
	fyne package --os darwin --id com.kristas.kuui --icon icon.png --release
release-android:
	fyne package --os web --id com.kristas.kuui --icon icon.png

build:
	go build -ldflags="-w -s"