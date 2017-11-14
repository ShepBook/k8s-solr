package k8s

import (
	"fmt"

	"github.com/previousnext/k8s-solr/crd"
)

func Name(solr *crd.Solr) string {
	return fmt.Sprintf("solr-%s", solr.ObjectMeta.Name)
}
