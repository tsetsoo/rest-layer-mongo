package mongo

import (
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/schema/query"
	"gopkg.in/mgo.v2/bson"
)

// getField translate a schema field into a MongoDB field:
//
//  - id -> _id with in order to tape on the mongo primary key
func getField(f string) string {
	if f == "id" {
		return "_id"
	}
	return f
}

// getQuery transform a query into a Mongo query
func getQuery(q *query.Query) (bson.M, error) {
	return translatePredicate(q.Predicate)
}

// getSort transform a resource.Lookup into a Mongo sort list.
// If the sort list is empty, fallback to _id.
func getSort(q *query.Query) []string {
	if len(q.Sort) == 0 {
		return []string{"_id"}
	}
	s := make([]string, len(q.Sort))
	for i, sort := range q.Sort {
		if sort.Reversed {
			s[i] = "-" + getField(sort.Name)
		} else {
			s[i] = getField(sort.Name)
		}
	}
	return s
}

func translatePredicate(q query.Predicate) (bson.M, error) {
	b := bson.M{}
	for _, exp := range q {
		switch t := exp.(type) {
		case query.And:
			s := []bson.M{}
			for _, subExp := range t {
				sb, err := translatePredicate(query.Predicate{subExp})
				if err != nil {
					return nil, err
				}
				s = append(s, sb)
			}
			b["$and"] = s
		case query.Or:
			s := []bson.M{}
			for _, subExp := range t {
				sb, err := translatePredicate(query.Predicate{subExp})
				if err != nil {
					return nil, err
				}
				s = append(s, sb)
			}
			b["$or"] = s
		case query.In:
			b[getField(t.Field)] = bson.M{"$in": valuesToInterface(t.Values)}
		case query.NotIn:
			b[getField(t.Field)] = bson.M{"$nin": valuesToInterface(t.Values)}
		case query.Equal:
			b[getField(t.Field)] = t.Value
		case query.NotEqual:
			b[getField(t.Field)] = bson.M{"$ne": t.Value}
		case query.GreaterThan:
			b[getField(t.Field)] = bson.M{"$gt": t.Value}
		case query.GreaterOrEqual:
			b[getField(t.Field)] = bson.M{"$gte": t.Value}
		case query.LowerThan:
			b[getField(t.Field)] = bson.M{"$lt": t.Value}
		case query.LowerOrEqual:
			b[getField(t.Field)] = bson.M{"$lte": t.Value}
		case query.Regex:
			b[getField(t.Field)] = bson.M{"$regex": t.Value.String()}
		default:
			return nil, resource.ErrNotImplemented
		}
	}
	return b, nil
}

func valuesToInterface(v []query.Value) []interface{} {
	I := make([]interface{}, len(v))
	for i, _v := range v {
		I[i] = _v
	}
	return I
}