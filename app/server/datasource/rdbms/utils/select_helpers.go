package utils

import (
	"fmt"
	"strings"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"

	api_service_protos "github.com/ydb-platform/fq-connector-go/api/service/protos"
	"github.com/ydb-platform/fq-connector-go/common"
)

func selectWhatToYDBColumns(selectWhat *api_service_protos.TSelect_TWhat) ([]*Ydb.Column, error) {
	var columns []*Ydb.Column

	for i, item := range selectWhat.Items {
		column := item.GetColumn()
		if column == nil {
			return nil, fmt.Errorf("item #%d (%v) is not a column", i, item)
		}

		columns = append(columns, column)
	}

	return columns, nil
}

func makeTSelectTWhatForEmptyColumnsRequest() *api_service_protos.TSelect_TWhat {
	return &api_service_protos.TSelect_TWhat{
		Items: []*api_service_protos.TSelect_TWhat_TItem{
			{
				Payload: &api_service_protos.TSelect_TWhat_TItem_Column{
					Column: &Ydb.Column{
						Name: "",
						Type: common.MakePrimitiveType(Ydb.Type_INT64),
					},
				},
			},
		},
	}
}

func formatSelectClause(
	formatter SQLFormatter,
	selectWhat *api_service_protos.TSelect_TWhat,
	fakeZeroOnEmptyColumnsSet bool,
) (string, *api_service_protos.TSelect_TWhat, error) {
	var sb strings.Builder

	columns, err := selectWhatToYDBColumns(selectWhat)
	if err != nil {
		return "", nil, fmt.Errorf("convert Select.What.Items to Ydb.Columns: %w", err)
	}

	var newSelectWhat *api_service_protos.TSelect_TWhat

	// for the case of empty column set select some constant for constructing a valid sql statement
	if len(columns) == 0 {
		if !fakeZeroOnEmptyColumnsSet {
			return "", nil, fmt.Errorf("empty columns set")
		}

		sb.WriteString("0")

		// YQ-3314: is needed only in select COUNT(*) for ydb datasource.
		// 		In PostgreSQL or ClickHouse type_mapper is based on typeNames that are extraceted from column.DatabaseTypeName().
		//		But in ydb type_mapper is based on ydbTypes, that are extracted from TSelect_TWhat
		newSelectWhat = makeTSelectTWhatForEmptyColumnsRequest()
	} else {
		for i, column := range columns {
			sb.WriteString(formatter.SanitiseIdentifier(column.GetName()))

			if i != len(columns)-1 {
				sb.WriteString(", ")
			}
		}

		newSelectWhat = selectWhat
	}

	return sb.String(), newSelectWhat, nil
}

// DefaultSelectQueryRender doesn't take into account splitting
func DefaultSelectQueryRender(parts *SelectQueryParts) (string, error) {
	var sb strings.Builder

	sb.WriteString("SELECT ")
	sb.WriteString(parts.SelectClause)
	sb.WriteString(" FROM ")
	sb.WriteString(parts.FromClause)

	if parts.WhereClause != "" {
		sb.WriteString(" WHERE ")
		sb.WriteString(parts.WhereClause)
	}

	return sb.String(), nil
}
