.PHONY: install simulator

install:
	@set -e; \
	export GOARCH=arm; \
	export GOOS=linux; \
	echo "Build"; \
	go build github.com/jypelle/vekigi/cmd/vekigisrv; \
	echo "Stop"; \
	ssh pi@vekigi "sudo systemctl stop vekigisrv"; \
	echo "Deploy"; \
	scp -p vekigisrv pi@vekigi:/home/pi/vekigisrv; \
	rm vekigisrv; \
	echo "Start"; \
	ssh pi@vekigi "sudo systemctl start vekigisrv";

simulator:
	go run ./cmd/vekigisrv -d -s run;
