package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/aws-sdk-go-v2/service/sfn/types"

	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

func tableAwsStepFunctionsStateMachineExecution(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_sfn_state_machine_execution",
		Description: "AWS Step Functions State Machine Execution",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("execution_arn"),
			IgnoreConfig: &plugin.IgnoreConfig{
				ShouldIgnoreErrorFunc: shouldIgnoreErrors([]string{"InvalidParameter", "ExecutionDoesNotExist", "InvalidArn"}),
			},
			Hydrate: getStepFunctionsStateMachineExecution,
			Tags:    map[string]string{"service": "states", "action": "DescribeExecution"},
		},
		List: &plugin.ListConfig{
			Hydrate:       listStepFunctionsStateMachineExecutions,
			Tags:          map[string]string{"service": "states", "action": "ListExecutions"},
			ParentHydrate: listStepFunctionsStateMachines,
			KeyColumns: []*plugin.KeyColumn{
				{Name: "status", Require: plugin.Optional},
				{Name: "state_machine_arn", Require: plugin.Optional},
			},
		},
		HydrateConfig: []plugin.HydrateConfig{
			{
				Func: getStepFunctionsStateMachineExecution,
				Tags: map[string]string{"service": "states", "action": "DescribeExecution"},
			},
		},
		GetMatrixItemFunc: SupportedRegionMatrix(AWS_STATES_SERVICE_ID),
		Columns: awsRegionalColumns([]*plugin.Column{
			{
				Name:        "name",
				Description: "The name of the execution.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "execution_arn",
				Description: "The Amazon Resource Name (ARN) that identifies the execution.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "status",
				Description: "The current status of the execution.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "input",
				Description: "The string that contains the JSON input data of the execution.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},
			{
				Name:        "cause",
				Description: "The cause string if the state machine execution failed.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},
			{
				Name:        "error",
				Description: "The error string if the state machine execution failed.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},
			{
				Name:        "map_run_arn",
				Description: "The Amazon Resource Name (ARN) that identifies a Map Run, which dispatched this execution.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},
			{
				Name:        "redrive_count",
				Description: "The number of times you've redriven an execution. If you have not yet redriven an execution, the redriveCount is 0.",
				Type:        proto.ColumnType_INT,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},
			{
				Name:        "redrive_date",
				Description: "The date the execution was last redriven.",
				Type:        proto.ColumnType_TIMESTAMP,
				Hydrate:     getStepFunctionsStateMachineExecution,
				Transform:   transform.FromField("RedriveDate").Transform(transform.NullIfZeroValue),
			},
			{
				Name:        "redrive_status",
				Description: "Indicates whether or not an execution can be redriven at a given point in time.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},
			{
				Name:        "redrive_status_reason",
				Description: "When redriveStatus is NOT_REDRIVABLE , redriveStatusReason specifies the reason why an execution cannot be redriven.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},
			{
				Name:        "state_machine_alias_arn",
				Description: "The Amazon Resource Name (ARN) of the state machine alias associated with the execution.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},
			{
				Name:        "state_machine_version_arn",
				Description: "The Amazon Resource Name (ARN) of the state machine version associated with the execution.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},
			{
				Name:        "output",
				Description: "The JSON output data of the execution.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},
			{
				Name:        "start_date",
				Description: "The date the execution started.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "state_machine_arn",
				Description: "The Amazon Resource Name (ARN) of the executed state machine.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "stop_date",
				Description: "If the execution already ended, the date the execution stopped.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			{
				Name:        "trace_header",
				Description: "The AWS X-Ray trace header that was passed to the execution.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},
			{
				Name:        "input_details",
				Description: "Provides details about execution input or output.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},
			{
				Name:        "output_details",
				Description: "Provides details about execution input or output.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getStepFunctionsStateMachineExecution,
			},

			// Standard columns for all tables
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "akas",
				Description: resourceInterfaceDescription("akas"),
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("ExecutionArn").Transform(transform.EnsureStringArray),
			},
		}),
	}
}

//// LIST FUNCTION

func listStepFunctionsStateMachineExecutions(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// Create session
	svc, err := StepFunctionsClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("aws_sfn_state_machine_execution.listStepFunctionsStateMachineExecutions", "connection_error", err)
		return nil, err
	}

	if svc == nil {
		// Unsupported region check
		return nil, nil
	}

	stateMachineArn := h.Item.(types.StateMachineListItem).StateMachineArn

	equalQuals := d.EqualsQuals
	// Minimize the API call with the given state machine ARN
	if equalQuals["state_machine_arn"] != nil {
		if equalQuals["state_machine_arn"].GetStringValue() != "" {
			if equalQuals["state_machine_arn"].GetStringValue() != "" && equalQuals["state_machine_arn"].GetStringValue() != *stateMachineArn {
				return nil, nil
			}
		} else if len(getListValues(equalQuals["state_machine_arn"].GetListValue())) > 0 {
			if !strings.Contains(fmt.Sprint(getListValues(equalQuals["state_machine_arn"].GetListValue())), *stateMachineArn) {
				return nil, nil
			}
		}
	}

	maxLimit := int32(1000)
	// If the requested number of items is less than the paging max limit
	// set the limit to that instead
	limit := d.QueryContext.Limit
	if d.QueryContext.Limit != nil {
		if *limit < int64(maxLimit) {
			maxLimit = int32(*limit)
		}
	}
	input := &sfn.ListExecutionsInput{
		StateMachineArn: stateMachineArn,
		MaxResults:      int32(maxLimit),
	}
	if equalQuals["status"] != nil {
		input.StatusFilter = types.ExecutionStatus(equalQuals["status"].GetStringValue())
	}
	paginator := sfn.NewListExecutionsPaginator(svc, input, func(o *sfn.ListExecutionsPaginatorOptions) {
		o.Limit = maxLimit
		o.StopOnDuplicateToken = true
	})

	for paginator.HasMorePages() {
		// apply rate limiting
		d.WaitForListRateLimit(ctx)

		output, err := paginator.NextPage(ctx)
		if err != nil {
			plugin.Logger(ctx).Error("aws_sfn_state_machine_execution.listStepFunctionsStateMachineExecutions", "api_error", err)
			return nil, err
		}
		for _, execution := range output.Executions {
			d.StreamListItem(ctx, execution)

			// Context may get cancelled due to manual cancellation or if the limit has been reached
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
	}

	if err != nil {
		plugin.Logger(ctx).Error("aws_sfn_state_machine_execution.listStepFunctionsStateMachineExecutions", "api_error", err)
		return nil, err
	}

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getStepFunctionsStateMachineExecution(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	var arn string
	if h.Item != nil {
		arn = *h.Item.(types.ExecutionListItem).ExecutionArn
	} else {
		arn = d.EqualsQuals["execution_arn"].GetStringValue()
	}

	// Create Session
	svc, err := StepFunctionsClient(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("aws_sfn_state_machine_execution.getStepFunctionsStateMachineExecution", "connection_error", err)
		return nil, err
	}

	if svc == nil {
		// Unsupported region check
		return nil, nil
	}

	// Build the params
	params := &sfn.DescribeExecutionInput{
		ExecutionArn: &arn,
	}

	// Get call
	data, err := svc.DescribeExecution(ctx, params)
	if err != nil {
		plugin.Logger(ctx).Error("aws_sfn_state_machine_execution.getStepFunctionsStateMachineExecution", "api_error", err)
		return nil, err
	}

	return data, nil
}
