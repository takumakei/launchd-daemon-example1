.DEFAULT_GOAL := help

.PHONY: help build lint install uninstall info start stop

LOGDIR := /var/log/example1

help: ## show this message
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-9s\033[0m %s\n", $$1, $$2}'

build: ## go build -trimpath -o example1 .
	go build -trimpath -o example1 .

example1:
	go build -trimpath -o $@ .

lint: ## plutil -lint com.takumakei.example1.plist
	plutil -lint com.takumakei.example1.plist

install: example1 com.takumakei.example1.plist
install: ## [sudo] install as a daemon
	sudo mkdir -p $(LOGDIR)
	sudo touch $(LOGDIR)/std{out,err}.log
	sudo chown -R daemon:daemon $(LOGDIR)
	sudo /usr/bin/install -o root -g wheel -m 644 com.takumakei.example1.plist /Library/LaunchDaemons/
	sudo /usr/bin/install -o daemon -g daemon -m 755 example1 /opt/
	sudo launchctl load -Fw /Library/LaunchDaemons/com.takumakei.example1.plist

uninstall: ## [sudo] uninstall the daemon
	sudo launchctl unload -Fw /Library/LaunchDaemons/com.takumakei.example1.plist
	sudo /bin/rm -fr /opt/example1
	sudo /bin/rm -fr /Library/LaunchDaemons/com.takumakei.example1.plist
	sudo /bin/rm -fr $(LOGDIR)

info: ## [sudo] launchctl list com.takumakei.example1
	sudo launchctl list com.takumakei.example1

start: ## [sudo] launchctl start com.takumakei.example1
	sudo launchctl start com.takumakei.example1

stop: ## [sudo] launchctl stop com.takumakei.example1
	sudo launchctl stop com.takumakei.example1

logf: ## tail -fq $(LOGDIR)/*.log
	tail -fq $(LOGDIR)/*.log
