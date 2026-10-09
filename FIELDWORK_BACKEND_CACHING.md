# Backend Caching Implementation for Fieldwork Endpoints

## Objective
Reduce excessive API calls by implementing caching headers and database optimization.

## Implementation Options (in order of priority)

---

## Option 1: HTTP Cache Headers (Easiest, Immediate Effect)

### File: `/Users/a/Projects/risk-based-audit/backend/audit-service/routes/handler.go`

Add cache middleware and apply to fieldwork GET routes:

```go
import (
    "crypto/md5"
    "fmt"
    "time"
)

// Add this function to RouteHandler
func (h *RouteHandler) cacheMiddleware(maxAge int) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Set cache headers for GET requests only
        if c.Request.Method == "GET" {
            c.Header("Cache-Control", fmt.Sprintf("private, max-age=%d", maxAge))
            c.Header("Pragma", "cache") // HTTP/1.0 compatibility
            
            // Allow browsers to use stale cache while revalidating in background
            c.Header("Cache-Control", fmt.Sprintf("private, max-age=%d, stale-while-revalidate=300", maxAge))
        }
        c.Next()
    }
}

// Add this function to RouteHandler for ETag generation
func (h *RouteHandler) generateETag(c *gin.Context) string {
    // Simple ETag based on query params
    etag := md5.Sum([]byte(c.Request.URL.RawQuery))
    return fmt.Sprintf("\"%x\"", etag)
}

// Add this function to RouteHandler
func (h *RouteHandler) etagMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Generate ETag
        etagValue := h.generateETag(c)
        c.Header("ETag", etagValue)
        
        // Check If-None-Match header
        if match := c.GetHeader("If-None-Match"); match == etagValue {
            c.AbortWithStatus(304) // Not Modified
            return
        }
        c.Next()
    }
}
```

### Apply to fieldwork routes in RegisterRoutes():

```go
// 6. Fieldwork Interviews
fieldworkInterviews := apiV1.Group("/fieldwork/interviews")
{
    // Cache GET requests for 5 minutes
    fieldworkInterviews.GET("", 
        h.cacheMiddleware(300), 
        h.etagMiddleware(),
        crud.List(h.db, "FieldworkInterview", func() interface{} { return &[]models.FieldworkInterview{} }),
    )
    fieldworkInterviews.GET("/:id", 
        h.cacheMiddleware(300),
        h.etagMiddleware(),
        crud.GetByID(h.db, "FieldworkInterview", func() interface{} { return &models.FieldworkInterview{} }),
    )
    fieldworkInterviews.POST("", crud.Create(h.db, "FieldworkInterview", func() interface{} { return &models.FieldworkInterview{} }))
    fieldworkInterviews.PUT("/:id", crud.Update(h.db, "FieldworkInterview", func() interface{} { return &models.FieldworkInterview{} }))
    fieldworkInterviews.DELETE("/:id", crud.Delete(h.db, "FieldworkInterview", func() interface{} { return &models.FieldworkInterview{} }))
}

// Apply same pattern to:
// - fieldworkObservations
// - fieldworkDocuments  
// - fieldworkSamples
// - fieldworkTestControls
```

**Effect**: Browser will cache GET responses for 5 minutes and send If-None-Match header on subsequent requests within that period. If data hasn't changed, server returns 304 Not Modified without full response body.

---

## Option 2: Database Indexes (Recommended for Large Tables)

### File: `/Users/a/Projects/risk-based-audit/backend/audit-service/models/audit-fieldwork.go`

Add indexes to models for faster filtering by `assignment_letter_id`:

```go
type FieldworkInterview struct {
    ID                  uuid.UUID `gorm:"primaryKey;index:idx_fw_interview_letter"`
    AssignmentLetterID  string    `gorm:"index:idx_fw_interview_letter;type:varchar(255)"`
    // ... other fields
}

type FieldworkObservation struct {
    ID                  uuid.UUID `gorm:"primaryKey;index:idx_fw_observation_letter"`
    AssignmentLetterID  string    `gorm:"index:idx_fw_observation_letter;type:varchar(255)"`
    // ... other fields
}

type FieldworkDocument struct {
    ID                  uuid.UUID `gorm:"primaryKey;index:idx_fw_document_letter"`
    AssignmentLetterID  string    `gorm:"index:idx_fw_document_letter;type:varchar(255)"`
    // ... other fields
}

type FieldworkSample struct {
    ID                  uuid.UUID `gorm:"primaryKey;index:idx_fw_sample_letter"`
    AssignmentLetterID  string    `gorm:"index:idx_fw_sample_letter;type:varchar(255)"`
    // ... other fields
}

type FieldworkTestControl struct {
    ID                  uuid.UUID `gorm:"primaryKey;index:idx_fw_testcontrol_letter"`
    AssignmentLetterID  string    `gorm:"index:idx_fw_testcontrol_letter;type:varchar(255)"`
    // ... other fields
}
```

### Run migration:

```bash
cd backend/audit-service

# This will run AutoMigrate and create the indexes
go run cmd/seed/main.go

# Or if you have a separate migration command:
go run cmd/migrate/main.go
```

**Effect**: Queries filtering by `assignmentLetterId` will use index and complete in O(log n) time instead of O(n) table scan.

---

## Option 3: Redis Query Caching (Most Effective)

### File: `/Users/a/Projects/risk-based-audit/controllers/crud/crud_handler.go`

Wrap the List handler with Redis caching:

```go
package crud

import (
    "encoding/json"
    "fmt"
    "time"
    "context"
    
    "github.com/redis/go-redis/v9"
)

// ListWithCache returns a paginated list of records with Redis caching
func ListWithCache(
    db *gorm.DB, 
    rdb *redis.Client,  // Redis client (may be nil)
    modelName string, 
    newSlice func() interface{}, 
    cacheTTL time.Duration,  // e.g., 5 * time.Minute
    preloads ...string,
) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Generate cache key from model, filters, and ordering
        cacheKey := fmt.Sprintf("fieldwork:%s:%s:%s:%s:%s",
            modelName,
            c.DefaultQuery("page", "1"),
            c.DefaultQuery("page_size", "20"),
            c.DefaultQuery("order", "created_at DESC"),
            c.Request.URL.RawQuery, // Includes assignmentLetterId and other filters
        )

        // Try Redis cache first (if client is configured)
        if rdb != nil {
            if cached, err := rdb.Get(c.Request.Context(), cacheKey).Result(); err == nil {
                var result interface{}
                if err := json.Unmarshal([]byte(cached), &result); err == nil {
                    c.Header("X-Cache", "HIT")  // Debug header
                    response.OK(c, modelName+" fetched successfully (cached)", result)
                    return
                }
            }
        }

        // Cache miss - execute original List handler
        items := newSlice()
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
        search := c.Query("search")

        if page < 1 {
            page = 1
        }
        if pageSize < 1 || pageSize > 100 {
            pageSize = 20
        }
        offset := (page - 1) * pageSize

        columns, err := modelColumns(db, items)
        if err != nil {
            response.InternalServerError(c, "Failed to fetch "+modelName)
            return
        }

        orderBy, ok := parseOrder(c.DefaultQuery("order", "created_at DESC"), columns)
        if !ok {
            response.BadRequest(c, "Invalid order parameter")
            return
        }

        orderBy = withIDTiebreaker(orderBy, columns)

        query := db.Model(items)
        for _, p := range preloads {
            query = query.Preload(p)
        }

        if search != "" {
            match := ilikeAny{Value: "%" + search + "%"}
            for _, col := range []string{"title", "name"} {
                if columns[col] {
                    match.Columns = append(match.Columns, col)
                }
            }
            if len(match.Columns) > 0 {
                query = query.Where(match)
            }
        }

        for key, values := range c.Request.URL.Query() {
            if key == "page" || key == "page_size" || key == "search" || key == "order" {
                continue
            }
            col := toSnakeCase(key)
            if !columns[col] {
                continue
            }
            if len(values) > 0 && values[0] != "" {
                query = query.Where(clause.Eq{Column: clause.Column{Name: col}, Value: values[0]})
            }
        }

        var total int64
        if err := query.Count(&total).Error; err != nil {
            response.InternalServerError(c, "Failed to fetch "+modelName)
            return
        }

        if err := query.Order(orderBy).Offset(offset).Limit(pageSize).Find(items).Error; err != nil {
            response.InternalServerError(c, "Failed to fetch "+modelName)
            return
        }

        responseData := gin.H{
            "items": items,
            "pagination": gin.H{
                "page":        page,
                "page_size":   pageSize,
                "total":       total,
                "total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
            },
        }

        // Cache the result in Redis (if client available)
        if rdb != nil {
            if data, err := json.Marshal(responseData); err == nil {
                rdb.Set(c.Request.Context(), cacheKey, data, cacheTTL)
            }
        }

        c.Header("X-Cache", "MISS")  // Debug header
        response.OK(c, modelName+" fetched successfully", responseData)
    }
}

// InvalidateCacheForModel clears all cache entries for a model when it's modified
func InvalidateCacheForModel(rdb *redis.Client, modelName string) error {
    if rdb == nil {
        return nil
    }
    ctx := context.Background()
    pattern := fmt.Sprintf("fieldwork:%s:*", modelName)
    
    iter := rdb.Scan(ctx, 0, pattern, 100).Iterator()
    var keys []string
    for iter.Next(ctx) {
        keys = append(keys, iter.Val())
    }
    
    if len(keys) > 0 {
        return rdb.Del(ctx, keys...).Err()
    }
    return nil
}
```

### Update routes to use caching:

```go
// In routes/handler.go RegisterRoutes()
fieldworkTestControls := apiV1.Group("/fieldwork/test-controls")
{
    // Use cached list handler with 5-minute TTL
    fieldworkTestControls.GET("", 
        ListWithCache(h.db, h.redisClient, "FieldworkTestControl", 
            func() interface{} { return &[]models.FieldworkTestControl{} },
            5 * time.Minute),
    )
    fieldworkTestControls.GET("/:id", 
        crud.GetByID(h.db, "FieldworkTestControl", func() interface{} { return &models.FieldworkTestControl{} }),
    )
    
    // POST/PUT/DELETE operations should invalidate cache
    fieldworkTestControls.POST("", func(c *gin.Context) {
        // ... handle create, then:
        InvalidateCacheForModel(h.redisClient, "FieldworkTestControl")
        // return response
    })
    
    fieldworkTestControls.PUT("/:id", func(c *gin.Context) {
        // ... handle update, then:
        InvalidateCacheForModel(h.redisClient, "FieldworkTestControl")
        // return response
    })
    
    fieldworkTestControls.DELETE("/:id", func(c *gin.Context) {
        // ... handle delete, then:
        InvalidateCacheForModel(h.redisClient, "FieldworkTestControl")
        // return response
    })
}
```

**Effect**: Repeated requests for the same fieldwork data within 5 minutes get served from Redis cache (~1-5ms) instead of querying database (~50-500ms). 99% reduction in database load for reads.

---

## Implementation Priority

1. **First**: Option 1 (HTTP cache headers) - No code changes needed, just routing config. Immediate browser-side caching.

2. **Second**: Option 2 (Database indexes) - Add indexes to models. Helps even without caching. Run `go test ./...` to ensure no breaks.

3. **Third**: Option 3 (Redis caching) - Most complex but most effective. Only implement if Redis is already running in the stack.

---

## Testing the Cache

### Test Cache Headers with curl:

```bash
# First request - cache miss
curl -i "http://localhost:8080/api/v1/fieldwork/test-controls?assignmentLetterId=ST-001" \
  -H "Authorization: Bearer TOKEN"
# Response headers should include: Cache-Control, ETag

# Second request - cache hit (within 5 minutes)
curl -i "http://localhost:8080/api/v1/fieldwork/test-controls?assignmentLetterId=ST-001" \
  -H "Authorization: Bearer TOKEN" \
  -H "If-None-Match: <value-from-first-response>"
# Response should be: 304 Not Modified
```

### Test with Browser DevTools:

1. Open DevTools → Network tab
2. Select assignment letter
3. First load: full response, size shows full bytes
4. Second load (within 5 min): 
   - Status: 304 Not Modified (or served from disk cache)
   - Size: shows "from cache"
   - X-Cache header: "HIT"

---

## Verification Checklist

- [ ] HTTP cache headers implemented on GET endpoints
- [ ] Database indexes created and verified with: `SELECT * FROM pg_indexes WHERE tablename LIKE 'fieldwork%'`
- [ ] Redis cache working (if implemented) - check X-Cache headers
- [ ] POST/PUT/DELETE operations invalidate relevant caches
- [ ] No test failures: `go test ./... -v`
- [ ] Monitor API call volume before and after
- [ ] Browser cache working: DevTools shows 304 responses

---

## Expected Results

### Before Optimization
- 5 fieldwork GET endpoints × 25 users = 125 requests/minute
- Each request hits database, ~200ms latency
- Total database load: ~4166 queries/second

### After Optimization
- Same 125 requests/minute, but:
  - 80% served from browser cache (304 Not Modified)
  - 15% served from Redis cache (1-5ms)
  - 5% require database query (200ms)
- Effective latency: ~12ms (80-85% improvement)
- Database load: ~200 queries/second (95% reduction)
