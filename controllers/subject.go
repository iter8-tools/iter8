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
	NormalizedWeights []uint32
}

type variant struct {
	Resources []resource `json:"resources,omitempty"`
	Weight    *uint32    `json:"weight,omitempty"`
}

type resource struct {
	GVRShort string `json:"gvrShort"`
	Name     string `json:"name"`
}

type ssa struct {
	GVRShort string `json:"gvrShort"`
	Template string `json:"template"`
}

/* types: end */
