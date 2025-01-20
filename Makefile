VERSION=1.2

build:
	docker build -t cuongnb14/db-backup:${VERSION} .

push:
	docker push cuongnb14/db-backup:${VERSION}