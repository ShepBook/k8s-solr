#!/usr/bin/make -f

IMAGE=previousnext/solr

cores:
	docker build -t $(IMAGE):init init
	docker build -t $(IMAGE):5.x 5.x

release: cores
	docker push $(IMAGE):init
	docker push $(IMAGE):5.x