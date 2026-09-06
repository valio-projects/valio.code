package catalog

import "github.com/valio-projects/valio.code/internal/domain/repositories"

// Repository aliases the domain persistence port used by this application feature.
type Repository = repositories.CatalogRepository
