#!/usr/bin/make -f

IMAGE=ShepBook/solr

cores:
	docker build -t $(IMAGE):init init
	docker build -t $(IMAGE):7.3 7.3

release: cores
	docker push $(IMAGE):init
	docker push $(IMAGE):7.3
