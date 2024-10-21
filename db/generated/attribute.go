// Code generated by SQLBoiler 4.16.2 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package dbgen

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// Attribute is an object representing the database table.
type Attribute struct {
	AssetID    int64  `boil:"asset_id" json:"asset_id" toml:"asset_id" yaml:"asset_id"`
	ID         int64  `boil:"id" json:"id" toml:"id" yaml:"id"`
	Subtype    string `boil:"subtype" json:"subtype" toml:"subtype" yaml:"subtype"`
	Name       string `boil:"name" json:"name" toml:"name" yaml:"name"`
	ProviderID string `boil:"provider_id" json:"provider_id" toml:"provider_id" yaml:"provider_id"`

	R *attributeR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L attributeL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var AttributeColumns = struct {
	AssetID    string
	ID         string
	Subtype    string
	Name       string
	ProviderID string
}{
	AssetID:    "asset_id",
	ID:         "id",
	Subtype:    "subtype",
	Name:       "name",
	ProviderID: "provider_id",
}

var AttributeTableColumns = struct {
	AssetID    string
	ID         string
	Subtype    string
	Name       string
	ProviderID string
}{
	AssetID:    "attribute.asset_id",
	ID:         "attribute.id",
	Subtype:    "attribute.subtype",
	Name:       "attribute.name",
	ProviderID: "attribute.provider_id",
}

// Generated where

var AttributeWhere = struct {
	AssetID    whereHelperint64
	ID         whereHelperint64
	Subtype    whereHelperstring
	Name       whereHelperstring
	ProviderID whereHelperstring
}{
	AssetID:    whereHelperint64{field: "\"open_bos\".\"attribute\".\"asset_id\""},
	ID:         whereHelperint64{field: "\"open_bos\".\"attribute\".\"id\""},
	Subtype:    whereHelperstring{field: "\"open_bos\".\"attribute\".\"subtype\""},
	Name:       whereHelperstring{field: "\"open_bos\".\"attribute\".\"name\""},
	ProviderID: whereHelperstring{field: "\"open_bos\".\"attribute\".\"provider_id\""},
}

// AttributeRels is where relationship names are stored.
var AttributeRels = struct {
	Asset string
}{
	Asset: "Asset",
}

// attributeR is where relationships are stored.
type attributeR struct {
	Asset *Asset `boil:"Asset" json:"Asset" toml:"Asset" yaml:"Asset"`
}

// NewStruct creates a new relationship struct
func (*attributeR) NewStruct() *attributeR {
	return &attributeR{}
}

func (r *attributeR) GetAsset() *Asset {
	if r == nil {
		return nil
	}
	return r.Asset
}

// attributeL is where Load methods for each relationship are stored.
type attributeL struct{}

var (
	attributeAllColumns            = []string{"asset_id", "id", "subtype", "name", "provider_id"}
	attributeColumnsWithoutDefault = []string{"subtype", "name", "provider_id"}
	attributeColumnsWithDefault    = []string{"asset_id", "id"}
	attributePrimaryKeyColumns     = []string{"id"}
	attributeGeneratedColumns      = []string{}
)

type (
	// AttributeSlice is an alias for a slice of pointers to Attribute.
	// This should almost always be used instead of []Attribute.
	AttributeSlice []*Attribute
	// AttributeHook is the signature for custom Attribute hook methods
	AttributeHook func(context.Context, boil.ContextExecutor, *Attribute) error

	attributeQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	attributeType                 = reflect.TypeOf(&Attribute{})
	attributeMapping              = queries.MakeStructMapping(attributeType)
	attributePrimaryKeyMapping, _ = queries.BindMapping(attributeType, attributeMapping, attributePrimaryKeyColumns)
	attributeInsertCacheMut       sync.RWMutex
	attributeInsertCache          = make(map[string]insertCache)
	attributeUpdateCacheMut       sync.RWMutex
	attributeUpdateCache          = make(map[string]updateCache)
	attributeUpsertCacheMut       sync.RWMutex
	attributeUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var attributeAfterSelectMu sync.Mutex
var attributeAfterSelectHooks []AttributeHook

var attributeBeforeInsertMu sync.Mutex
var attributeBeforeInsertHooks []AttributeHook
var attributeAfterInsertMu sync.Mutex
var attributeAfterInsertHooks []AttributeHook

var attributeBeforeUpdateMu sync.Mutex
var attributeBeforeUpdateHooks []AttributeHook
var attributeAfterUpdateMu sync.Mutex
var attributeAfterUpdateHooks []AttributeHook

var attributeBeforeDeleteMu sync.Mutex
var attributeBeforeDeleteHooks []AttributeHook
var attributeAfterDeleteMu sync.Mutex
var attributeAfterDeleteHooks []AttributeHook

var attributeBeforeUpsertMu sync.Mutex
var attributeBeforeUpsertHooks []AttributeHook
var attributeAfterUpsertMu sync.Mutex
var attributeAfterUpsertHooks []AttributeHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Attribute) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range attributeAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Attribute) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range attributeBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Attribute) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range attributeAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Attribute) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range attributeBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Attribute) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range attributeAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Attribute) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range attributeBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Attribute) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range attributeAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Attribute) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range attributeBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Attribute) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range attributeAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddAttributeHook registers your hook function for all future operations.
func AddAttributeHook(hookPoint boil.HookPoint, attributeHook AttributeHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		attributeAfterSelectMu.Lock()
		attributeAfterSelectHooks = append(attributeAfterSelectHooks, attributeHook)
		attributeAfterSelectMu.Unlock()
	case boil.BeforeInsertHook:
		attributeBeforeInsertMu.Lock()
		attributeBeforeInsertHooks = append(attributeBeforeInsertHooks, attributeHook)
		attributeBeforeInsertMu.Unlock()
	case boil.AfterInsertHook:
		attributeAfterInsertMu.Lock()
		attributeAfterInsertHooks = append(attributeAfterInsertHooks, attributeHook)
		attributeAfterInsertMu.Unlock()
	case boil.BeforeUpdateHook:
		attributeBeforeUpdateMu.Lock()
		attributeBeforeUpdateHooks = append(attributeBeforeUpdateHooks, attributeHook)
		attributeBeforeUpdateMu.Unlock()
	case boil.AfterUpdateHook:
		attributeAfterUpdateMu.Lock()
		attributeAfterUpdateHooks = append(attributeAfterUpdateHooks, attributeHook)
		attributeAfterUpdateMu.Unlock()
	case boil.BeforeDeleteHook:
		attributeBeforeDeleteMu.Lock()
		attributeBeforeDeleteHooks = append(attributeBeforeDeleteHooks, attributeHook)
		attributeBeforeDeleteMu.Unlock()
	case boil.AfterDeleteHook:
		attributeAfterDeleteMu.Lock()
		attributeAfterDeleteHooks = append(attributeAfterDeleteHooks, attributeHook)
		attributeAfterDeleteMu.Unlock()
	case boil.BeforeUpsertHook:
		attributeBeforeUpsertMu.Lock()
		attributeBeforeUpsertHooks = append(attributeBeforeUpsertHooks, attributeHook)
		attributeBeforeUpsertMu.Unlock()
	case boil.AfterUpsertHook:
		attributeAfterUpsertMu.Lock()
		attributeAfterUpsertHooks = append(attributeAfterUpsertHooks, attributeHook)
		attributeAfterUpsertMu.Unlock()
	}
}

// OneG returns a single attribute record from the query using the global executor.
func (q attributeQuery) OneG(ctx context.Context) (*Attribute, error) {
	return q.One(ctx, boil.GetContextDB())
}

// One returns a single attribute record from the query.
func (q attributeQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Attribute, error) {
	o := &Attribute{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbgen: failed to execute a one query for attribute")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// AllG returns all Attribute records from the query using the global executor.
func (q attributeQuery) AllG(ctx context.Context) (AttributeSlice, error) {
	return q.All(ctx, boil.GetContextDB())
}

// All returns all Attribute records from the query.
func (q attributeQuery) All(ctx context.Context, exec boil.ContextExecutor) (AttributeSlice, error) {
	var o []*Attribute

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "dbgen: failed to assign all query results to Attribute slice")
	}

	if len(attributeAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// CountG returns the count of all Attribute records in the query using the global executor
func (q attributeQuery) CountG(ctx context.Context) (int64, error) {
	return q.Count(ctx, boil.GetContextDB())
}

// Count returns the count of all Attribute records in the query.
func (q attributeQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: failed to count attribute rows")
	}

	return count, nil
}

// ExistsG checks if the row exists in the table using the global executor.
func (q attributeQuery) ExistsG(ctx context.Context) (bool, error) {
	return q.Exists(ctx, boil.GetContextDB())
}

// Exists checks if the row exists in the table.
func (q attributeQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "dbgen: failed to check if attribute exists")
	}

	return count > 0, nil
}

// Asset pointed to by the foreign key.
func (o *Attribute) Asset(mods ...qm.QueryMod) assetQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.AssetID),
	}

	queryMods = append(queryMods, mods...)

	return Assets(queryMods...)
}

// LoadAsset allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (attributeL) LoadAsset(ctx context.Context, e boil.ContextExecutor, singular bool, maybeAttribute interface{}, mods queries.Applicator) error {
	var slice []*Attribute
	var object *Attribute

	if singular {
		var ok bool
		object, ok = maybeAttribute.(*Attribute)
		if !ok {
			object = new(Attribute)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeAttribute)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeAttribute))
			}
		}
	} else {
		s, ok := maybeAttribute.(*[]*Attribute)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeAttribute)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeAttribute))
			}
		}
	}

	args := make(map[interface{}]struct{})
	if singular {
		if object.R == nil {
			object.R = &attributeR{}
		}
		args[object.AssetID] = struct{}{}

	} else {
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &attributeR{}
			}

			args[obj.AssetID] = struct{}{}

		}
	}

	if len(args) == 0 {
		return nil
	}

	argsSlice := make([]interface{}, len(args))
	i := 0
	for arg := range args {
		argsSlice[i] = arg
		i++
	}

	query := NewQuery(
		qm.From(`open_bos.asset`),
		qm.WhereIn(`open_bos.asset.id in ?`, argsSlice...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Asset")
	}

	var resultSlice []*Asset
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Asset")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for asset")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for asset")
	}

	if len(assetAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Asset = foreign
		if foreign.R == nil {
			foreign.R = &assetR{}
		}
		foreign.R.Attributes = append(foreign.R.Attributes, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.AssetID == foreign.ID {
				local.R.Asset = foreign
				if foreign.R == nil {
					foreign.R = &assetR{}
				}
				foreign.R.Attributes = append(foreign.R.Attributes, local)
				break
			}
		}
	}

	return nil
}

// SetAssetG of the attribute to the related item.
// Sets o.R.Asset to related.
// Adds o to related.R.Attributes.
// Uses the global database handle.
func (o *Attribute) SetAssetG(ctx context.Context, insert bool, related *Asset) error {
	return o.SetAsset(ctx, boil.GetContextDB(), insert, related)
}

// SetAsset of the attribute to the related item.
// Sets o.R.Asset to related.
// Adds o to related.R.Attributes.
func (o *Attribute) SetAsset(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Asset) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"open_bos\".\"attribute\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"asset_id"}),
		strmangle.WhereClause("\"", "\"", 2, attributePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.AssetID = related.ID
	if o.R == nil {
		o.R = &attributeR{
			Asset: related,
		}
	} else {
		o.R.Asset = related
	}

	if related.R == nil {
		related.R = &assetR{
			Attributes: AttributeSlice{o},
		}
	} else {
		related.R.Attributes = append(related.R.Attributes, o)
	}

	return nil
}

// Attributes retrieves all the records using an executor.
func Attributes(mods ...qm.QueryMod) attributeQuery {
	mods = append(mods, qm.From("\"open_bos\".\"attribute\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"open_bos\".\"attribute\".*"})
	}

	return attributeQuery{q}
}

// FindAttributeG retrieves a single record by ID.
func FindAttributeG(ctx context.Context, iD int64, selectCols ...string) (*Attribute, error) {
	return FindAttribute(ctx, boil.GetContextDB(), iD, selectCols...)
}

// FindAttribute retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindAttribute(ctx context.Context, exec boil.ContextExecutor, iD int64, selectCols ...string) (*Attribute, error) {
	attributeObj := &Attribute{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"open_bos\".\"attribute\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, attributeObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "dbgen: unable to select from attribute")
	}

	if err = attributeObj.doAfterSelectHooks(ctx, exec); err != nil {
		return attributeObj, err
	}

	return attributeObj, nil
}

// InsertG a single record. See Insert for whitelist behavior description.
func (o *Attribute) InsertG(ctx context.Context, columns boil.Columns) error {
	return o.Insert(ctx, boil.GetContextDB(), columns)
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Attribute) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("dbgen: no attribute provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(attributeColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	attributeInsertCacheMut.RLock()
	cache, cached := attributeInsertCache[key]
	attributeInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			attributeAllColumns,
			attributeColumnsWithDefault,
			attributeColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(attributeType, attributeMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(attributeType, attributeMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"open_bos\".\"attribute\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"open_bos\".\"attribute\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "dbgen: unable to insert into attribute")
	}

	if !cached {
		attributeInsertCacheMut.Lock()
		attributeInsertCache[key] = cache
		attributeInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// UpdateG a single Attribute record using the global executor.
// See Update for more documentation.
func (o *Attribute) UpdateG(ctx context.Context, columns boil.Columns) (int64, error) {
	return o.Update(ctx, boil.GetContextDB(), columns)
}

// Update uses an executor to update the Attribute.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Attribute) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	attributeUpdateCacheMut.RLock()
	cache, cached := attributeUpdateCache[key]
	attributeUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			attributeAllColumns,
			attributePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("dbgen: unable to update attribute, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"open_bos\".\"attribute\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, attributePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(attributeType, attributeMapping, append(wl, attributePrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: unable to update attribute row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: failed to get rows affected by update for attribute")
	}

	if !cached {
		attributeUpdateCacheMut.Lock()
		attributeUpdateCache[key] = cache
		attributeUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAllG updates all rows with the specified column values.
func (q attributeQuery) UpdateAllG(ctx context.Context, cols M) (int64, error) {
	return q.UpdateAll(ctx, boil.GetContextDB(), cols)
}

// UpdateAll updates all rows with the specified column values.
func (q attributeQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: unable to update all for attribute")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: unable to retrieve rows affected for attribute")
	}

	return rowsAff, nil
}

// UpdateAllG updates all rows with the specified column values.
func (o AttributeSlice) UpdateAllG(ctx context.Context, cols M) (int64, error) {
	return o.UpdateAll(ctx, boil.GetContextDB(), cols)
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o AttributeSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("dbgen: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), attributePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"open_bos\".\"attribute\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, attributePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: unable to update all in attribute slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: unable to retrieve rows affected all in update all attribute")
	}
	return rowsAff, nil
}

// UpsertG attempts an insert, and does an update or ignore on conflict.
func (o *Attribute) UpsertG(ctx context.Context, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns, opts ...UpsertOptionFunc) error {
	return o.Upsert(ctx, boil.GetContextDB(), updateOnConflict, conflictColumns, updateColumns, insertColumns, opts...)
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Attribute) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns, opts ...UpsertOptionFunc) error {
	if o == nil {
		return errors.New("dbgen: no attribute provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(attributeColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	attributeUpsertCacheMut.RLock()
	cache, cached := attributeUpsertCache[key]
	attributeUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, _ := insertColumns.InsertColumnSet(
			attributeAllColumns,
			attributeColumnsWithDefault,
			attributeColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			attributeAllColumns,
			attributePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("dbgen: unable to upsert attribute, could not build update column list")
		}

		ret := strmangle.SetComplement(attributeAllColumns, strmangle.SetIntersect(insert, update))

		conflict := conflictColumns
		if len(conflict) == 0 && updateOnConflict && len(update) != 0 {
			if len(attributePrimaryKeyColumns) == 0 {
				return errors.New("dbgen: unable to upsert attribute, could not build conflict column list")
			}

			conflict = make([]string, len(attributePrimaryKeyColumns))
			copy(conflict, attributePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"open_bos\".\"attribute\"", updateOnConflict, ret, update, conflict, insert, opts...)

		cache.valueMapping, err = queries.BindMapping(attributeType, attributeMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(attributeType, attributeMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "dbgen: unable to upsert attribute")
	}

	if !cached {
		attributeUpsertCacheMut.Lock()
		attributeUpsertCache[key] = cache
		attributeUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// DeleteG deletes a single Attribute record.
// DeleteG will match against the primary key column to find the record to delete.
func (o *Attribute) DeleteG(ctx context.Context) (int64, error) {
	return o.Delete(ctx, boil.GetContextDB())
}

// Delete deletes a single Attribute record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Attribute) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("dbgen: no Attribute provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), attributePrimaryKeyMapping)
	sql := "DELETE FROM \"open_bos\".\"attribute\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: unable to delete from attribute")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: failed to get rows affected by delete for attribute")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

func (q attributeQuery) DeleteAllG(ctx context.Context) (int64, error) {
	return q.DeleteAll(ctx, boil.GetContextDB())
}

// DeleteAll deletes all matching rows.
func (q attributeQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("dbgen: no attributeQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: unable to delete all from attribute")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: failed to get rows affected by deleteall for attribute")
	}

	return rowsAff, nil
}

// DeleteAllG deletes all rows in the slice.
func (o AttributeSlice) DeleteAllG(ctx context.Context) (int64, error) {
	return o.DeleteAll(ctx, boil.GetContextDB())
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o AttributeSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(attributeBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), attributePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"open_bos\".\"attribute\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, attributePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: unable to delete all from attribute slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "dbgen: failed to get rows affected by deleteall for attribute")
	}

	if len(attributeAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// ReloadG refetches the object from the database using the primary keys.
func (o *Attribute) ReloadG(ctx context.Context) error {
	if o == nil {
		return errors.New("dbgen: no Attribute provided for reload")
	}

	return o.Reload(ctx, boil.GetContextDB())
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *Attribute) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindAttribute(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAllG refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *AttributeSlice) ReloadAllG(ctx context.Context) error {
	if o == nil {
		return errors.New("dbgen: empty AttributeSlice provided for reload all")
	}

	return o.ReloadAll(ctx, boil.GetContextDB())
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *AttributeSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := AttributeSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), attributePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"open_bos\".\"attribute\".* FROM \"open_bos\".\"attribute\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, attributePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "dbgen: unable to reload all in AttributeSlice")
	}

	*o = slice

	return nil
}

// AttributeExistsG checks if the Attribute row exists.
func AttributeExistsG(ctx context.Context, iD int64) (bool, error) {
	return AttributeExists(ctx, boil.GetContextDB(), iD)
}

// AttributeExists checks if the Attribute row exists.
func AttributeExists(ctx context.Context, exec boil.ContextExecutor, iD int64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"open_bos\".\"attribute\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "dbgen: unable to check if attribute exists")
	}

	return exists, nil
}

// Exists checks if the Attribute row exists.
func (o *Attribute) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return AttributeExists(ctx, exec, o.ID)
}
