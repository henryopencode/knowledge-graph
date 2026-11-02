package content

import (
	"context"
	"sort"
)

type PublicationResult struct {
	ID       string `json:"id"`
	Revision string `json:"revision"`
}

// PublishAll is an explicit maintenance operation, never a startup default.
// Validate the full set before changing any policies or publication snapshots.
func (s *Store) PublishAll(ctx context.Context) ([]PublicationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	documents, err := s.readAll()
	if err != nil {
		return nil, err
	}
	groups := groupDocuments(documents)
	ids := []string{}
	for id := range groups {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	plans := []publicationPlan{}
	active := []published{}
	previous := map[string]bool{}
	results := []PublicationResult{}
	for _, id := range ids {
		variants := groups[id]
		previous[id] = variants[0].Channels.Blog
		for i := range variants {
			variants[i].Channels.Blog = true
			active = append(active, published{Document: variants[i]})
		}
		plan, err := s.plan(variants)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
		results = append(results, PublicationResult{ID: id, Revision: plan.Revision})
	}
	if err = validatePublicSlugs(active); err != nil {
		return nil, err
	}
	changed := []string{}
	restore := func() {
		for _, id := range changed {
			_ = s.writePolicy(id, previous[id])
		}
	}
	for _, id := range ids {
		if err = s.writePolicy(id, true); err != nil {
			restore()
			return nil, err
		}
		changed = append(changed, id)
	}
	if err = s.commitPublications(ctx, plans); err != nil {
		restore()
		return nil, err
	}
	return results, nil
}
