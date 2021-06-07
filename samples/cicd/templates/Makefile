ifndef VERSION
VERSION = $(shell git rev-parse --short head)
endif

candidate:
	cat deployment-candidate.yaml | sed "s/VERSION/v$(VERSION)/g" > ../deployment-candidate.yaml
	sed "s|  name: gitops-exp-|  name: gitops-exp-$(shell echo $$RANDOM)|" experiment.yaml > ../experiment.yaml
	#sed "s|  name: fortio-|  name: fortio-$(shell echo $$RANDOM)|" fortio.yaml > ../fortio.yaml
	

clean:
	rm -f ../deployment-candidate.yaml ../experiment.yaml ../fortio.yaml
