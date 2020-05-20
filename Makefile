all:
	go build -o bing-wallpaper bing-wallpaper.go

.PHONY: clean
clean:
	rm bing-wallpaper
