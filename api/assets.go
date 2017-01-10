package api

import (
	"errors"
	"net/http"

	"github.com/murdinc/awsm/aws"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
)

func getAssets(w http.ResponseWriter, r *http.Request) {
	// Get the listType
	assetType := chi.URLParam(r, "assetType")

	var resp interface{}
	var errs []error // multi-region
	var err error    // single region

	switch assetType {

	case "addresses":
		resp, errs = aws.GetAddresses("", false)

	case "alarms":
		resp, errs = aws.GetAlarms()

	case "autoscalegroups":
		resp, errs = aws.GetAutoScaleGroups("")

	case "iaminstanceprofiles":
		resp, err = aws.GetIAMInstanceProfiles("")

	case "iamroles":
		resp, err = aws.GetIAMRoles("")

	case "iamusers":
		resp, err = aws.GetIAMUsers("")

	case "images":
		resp, errs = aws.GetImages("", false)

	case "instances":
		resp, errs = aws.GetInstances("", false)

	case "keypairs":
		resp, errs = aws.GetKeyPairs("")

	case "launchconfigurations":
		resp, errs = aws.GetLaunchConfigurations("")

	case "loadbalancers":
		resp, errs = aws.GetLoadBalancers()

	case "scalingpolicies":
		resp, errs = aws.GetScalingPolicies()

	case "securitygroups":
		resp, errs = aws.GetSecurityGroups("")

	case "simpledbdomains":
		resp, errs = aws.GetSimpleDBDomains("")

	case "snapshots":
		resp, errs = aws.GetSnapshots("", false)

	case "subnets":
		resp, errs = aws.GetSubnets("")

	case "volumes":
		resp, errs = aws.GetVolumes("", false)

	case "vpcs":
		resp, errs = aws.GetVpcs("")

		/*
			case "buckets": // TODO
				resp, errs = aws.GetBuckets()
		*/

	default:
		err = errors.New("Unknown list type")
	}

	// Combine errors
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		render.JSON(w, r, map[string]interface{}{"assetType": assetType, "assets": resp, "success": true})
	} else {

		errStrs := make([]string, len(errs))

		for i, e := range errs {
			errStrs[i] = e.Error()
		}

		render.JSON(w, r, map[string]interface{}{"assetType": assetType, "assets": resp, "success": false, "errors": errStrs})
	}
}
