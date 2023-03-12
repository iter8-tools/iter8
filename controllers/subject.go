package controllers

import (
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/* types: begin */

type subject struct {
	// Todo: prune this down to agra.ObjectMeta instead of metav1.ObjectMeta
	mutex             sync.RWMutex
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Variants          []variant      `json:"variants,omitempty"`
	SSAs              map[string]ssa `json:"ssas,omitempty"`
	weights           []uint32
	normalizedWeights []uint32
}

type variant struct {
	Resources []resource `json:"resources,omitempty"`
	Weight    *uint32    `json:"resources,omitempty"`
}

type resource struct {
	GVKRShort string `json:"gvkrShort"`
	Name      string `json:"name"`
}

type ssa struct {
	GVKRShort string `json:"gvkrShort"`
	Template  string `json:"template"`
}

/* types: end */
