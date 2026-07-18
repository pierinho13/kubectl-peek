package usage

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

func ResourceAvailable(
	client discovery.DiscoveryInterface,
	gvr schema.GroupVersionResource,
) (bool, error) {
	return discovery.IsResourceEnabled(client, gvr)
}
