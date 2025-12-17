package relationship

import (
	"maps"
	"strings"
	"sync"

	"github.com/griffnb/core/lib/tools"
)

type Relationship struct {
	Type    string
	Field   string
	Package string
}

type relationshipRegistry struct {
	mut           sync.RWMutex
	Relationships map[string][]*Relationship
}

var (
	registry *relationshipRegistry
	once     sync.Once
)

func Registry() *relationshipRegistry {
	once.Do(func() {
		registry = &relationshipRegistry{
			Relationships: make(map[string][]*Relationship),
		}
	})
	return registry
}

func (r *relationshipRegistry) Register(packageName string, structPtr any) {
	fields := tools.ExtractFields(structPtr)

	for _, field := range fields {
		fk := field.Tag.Get("fk")
		if tools.Empty(fk) {
			continue
		}

		parts := strings.Split(fk, ":")
		if len(parts) != 2 {
			continue
		}

		relationshipType := parts[0]
		targetPackage := parts[1]

		fieldName := field.Tag.Get("column")

		r.register(packageName, fieldName, relationshipType, targetPackage)

	}
}

func (r *relationshipRegistry) register(packageName, field, relationshipType, targetPackage string) {
	r.mut.Lock()
	defer r.mut.Unlock()

	switch relationshipType {
	case "has_many", "has_one":
		if _, ok := r.Relationships[packageName]; !ok {
			r.Relationships[packageName] = make([]*Relationship, 0)
		}
		r.Relationships[packageName] = append(r.Relationships[packageName], &Relationship{
			Type:    relationshipType,
			Field:   field,
			Package: targetPackage,
		})
	case "belongs_to": // inverse relationship
		if _, ok := r.Relationships[targetPackage]; !ok {
			r.Relationships[targetPackage] = make([]*Relationship, 0)
		}
		r.Relationships[targetPackage] = append(r.Relationships[targetPackage], &Relationship{
			Type:    relationshipType,
			Field:   field,
			Package: packageName,
		})
	}
}

func (r *relationshipRegistry) Get(packageName string) []*Relationship {
	r.mut.RLock()
	defer r.mut.RUnlock()
	return r.Relationships[packageName]
}

func (r *relationshipRegistry) GetAll() map[string][]*Relationship {
	r.mut.RLock()
	defer r.mut.RUnlock()

	return maps.Clone(r.Relationships)
}
