package bonsai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/omc/bonsai-api-go/v1/bonsai"
)

func (s *ClientMockTestSuite) TestClusterClient_All() {
	s.serveMux.Get(bonsai.ClusterAPIBasePath, func(w http.ResponseWriter, _ *http.Request) {
		respStr := `
		{	
			"pagination": {		
				"page_number": 1,		
				"page_size": 10,
				"total_records": 3
			},	
			"clusters": [		
				{
					"slug": "first-testing-cluste-1234567890",
					"name": "first_testing_cluster",			
					"uri": "https://api.bonsai.io/clusters/first-testing-cluste-1234567890",			
					"plan": {		 	 
						"slug": "sandbox-aws-us-east-1",			  
						"uri": "https://api.bonsai.io/plans/sandbox-aws-us-east-1"			
					},	 	 
					"release": {			  
						"version": "7.2.0",			  
						"slug": "elasticsearch-7.2.0",	
						"package_name": "7.2.0",	
						"service_type": "elasticsearch",			  
						"uri": "https://api.bonsai.io/releases/elasticsearch-7.2.0"			
					},  		
					"space": {			  
						"path": "omc/bonsai/us-east-1/common",			  
						"region": "aws-us-east-1",			  
						"uri": "https://api.bonsai.io/spaces/omc/bonsai/us-east-1/common"			
					},	  	
					"stats": {			  
						"docs": 0,			  
						"shards_used": 0,			  
						"data_bytes_used": 0			
					},	    
					"access": {			  
						"host": "first-testing-cluste-1234567890.us-east-1.bonsaisearch.net",			  
						"port": 443,			 
						"scheme": "https"			
					},	  	
					"state": "PROVISIONED"		
				}, 	 
				{			
					"slug": "second-testing-clust-1234567890",			
					"name": "second_testing_cluster",			
					"uri": "https://api.bonsai.io/clusters/second-testing-clust-1234567890",			
					"plan": {		 	 
						"slug": "sandbox-aws-us-east-1",			  
						"uri": "https://api.bonsai.io/plans/sandbox-aws-us-east-1"			
					},	 	 
					"release": {			  
						"version": "7.2.0",			  
						"slug": "elasticsearch-7.2.0",	
						"package_name": "7.2.0",	
						"service_type": "elasticsearch",			  
						"uri": "https://api.bonsai.io/releases/elasticsearch-7.2.0"			
					},  		
					"space": {			  
						"path": "omc/bonsai/us-east-1/common",			  
						"region": "aws-us-east-1",			  
						"uri": "https://api.bonsai.io/spaces/omc/bonsai/us-east-1/common"			
					},	  	
					"stats": {			  
						"docs": 0,			  
						"shards_used": 0,			  
						"data_bytes_used": 0			
						},	    
					"access": {			  
						"host": "second-testing-clust-1234567890.us-east-1.bonsaisearch.net",			  
						"port": 443,			  
						"scheme": "https"			
					},	  	
					"state": "PROVISIONED"		
				},
				{
					"slug": "third-testing-clust-1234567890",			
					"name": "third_testing_cluster",			
					"uri": "https://api.bonsai.io/clusters/third-testing-clust-1234567890",			
					"plan": {		 	 
						"slug": "sandbox-aws-us-east-1",			  
						"uri": "https://api.bonsai.io/plans/sandbox-aws-us-east-1"			
					},	 	 
					"release": {			  
						"version": "7.2.0",			  
						"slug": "elasticsearch-7.2.0",	
						"package_name": "7.2.0",	
						"service_type": "elasticsearch",			  
						"uri": "https://api.bonsai.io/releases/elasticsearch-7.2.0"			
					},  		
					"space": {			  
						"path": "omc/bonsai/us-east-1/common",			  
						"region": "aws-us-east-1",			  
						"uri": "https://api.bonsai.io/spaces/omc/bonsai/us-east-1/common"			
					},	  	
					"stats": {			  
						"docs": 1500000,			  
						"shards_used": 14,			  
						"data_bytes_used": 93180912390			
						},	    
					"access": {			  
						"host": "third-testing-clust-1234567890.us-east-1.bonsaisearch.net",			  
						"port": 443,			  
						"scheme": "https"			
					},	  	
					"state": "PROVISIONED"		
				}
			]
		}
		`

		resp := &bonsai.ClustersResultList{Clusters: make([]bonsai.Cluster, 0, 2)}
		err := json.Unmarshal([]byte(respStr), resp)
		s.NoError(err, "unmarshal json into bonsai.ClustersResultList")

		err = json.NewEncoder(w).Encode(resp)
		s.NoError(err, "encode bonsai.ClustersResultList into json")
	})

	expect := []bonsai.Cluster{
		{
			Slug: "first-testing-cluste-1234567890",
			Name: "first_testing_cluster",
			URI:  "https://api.bonsai.io/clusters/first-testing-cluste-1234567890",
			Plan: bonsai.Plan{
				Slug:              "sandbox-aws-us-east-1",
				AvailableReleases: []bonsai.Release{},
				AvailableSpaces:   []bonsai.Space{},
				URI:               "https://api.bonsai.io/plans/sandbox-aws-us-east-1",
			},
			Release: bonsai.Release{
				Version:     "7.2.0",
				Slug:        "elasticsearch-7.2.0",
				PackageName: "7.2.0",
				ServiceType: "elasticsearch",
				URI:         "https://api.bonsai.io/releases/elasticsearch-7.2.0",
			},
			Space: bonsai.Space{
				Path:   "omc/bonsai/us-east-1/common",
				Region: "aws-us-east-1",
				URI:    "https://api.bonsai.io/spaces/omc/bonsai/us-east-1/common",
			},
			Stats: bonsai.ClusterStats{
				Docs:          0,
				ShardsUsed:    0,
				DataBytesUsed: 0,
			},
			Access: bonsai.ClusterAccess{
				Host:   "first-testing-cluste-1234567890.us-east-1.bonsaisearch.net",
				Port:   443,
				Scheme: "https",
			},
			State: bonsai.ClusterStateProvisioned,
		},
		{
			Slug: "second-testing-clust-1234567890",
			Name: "second_testing_cluster",
			URI:  "https://api.bonsai.io/clusters/second-testing-clust-1234567890",
			Plan: bonsai.Plan{
				Slug:              "sandbox-aws-us-east-1",
				AvailableReleases: []bonsai.Release{},
				AvailableSpaces:   []bonsai.Space{},
				URI:               "https://api.bonsai.io/plans/sandbox-aws-us-east-1",
			},
			Release: bonsai.Release{
				Version:     "7.2.0",
				Slug:        "elasticsearch-7.2.0",
				PackageName: "7.2.0",
				ServiceType: "elasticsearch",
				URI:         "https://api.bonsai.io/releases/elasticsearch-7.2.0",
			},
			Space: bonsai.Space{
				Path:   "omc/bonsai/us-east-1/common",
				Region: "aws-us-east-1",
				URI:    "https://api.bonsai.io/spaces/omc/bonsai/us-east-1/common",
			},
			Stats: bonsai.ClusterStats{
				Docs:          0,
				ShardsUsed:    0,
				DataBytesUsed: 0,
			},
			Access: bonsai.ClusterAccess{
				Host:   "second-testing-clust-1234567890.us-east-1.bonsaisearch.net",
				Port:   443,
				Scheme: "https",
			},
			State: bonsai.ClusterStateProvisioned,
		},
		{
			Slug: "third-testing-clust-1234567890",
			Name: "third_testing_cluster",
			URI:  "https://api.bonsai.io/clusters/third-testing-clust-1234567890",
			Plan: bonsai.Plan{
				Slug:              "sandbox-aws-us-east-1",
				AvailableReleases: []bonsai.Release{},
				AvailableSpaces:   []bonsai.Space{},
				URI:               "https://api.bonsai.io/plans/sandbox-aws-us-east-1",
			},
			Release: bonsai.Release{
				Version:     "7.2.0",
				Slug:        "elasticsearch-7.2.0",
				PackageName: "7.2.0",
				ServiceType: "elasticsearch",
				URI:         "https://api.bonsai.io/releases/elasticsearch-7.2.0",
			},
			Space: bonsai.Space{
				Path:   "omc/bonsai/us-east-1/common",
				Region: "aws-us-east-1",
				URI:    "https://api.bonsai.io/spaces/omc/bonsai/us-east-1/common",
			},
			Stats: bonsai.ClusterStats{
				Docs:          1500000,
				ShardsUsed:    14,
				DataBytesUsed: 93180912390,
			},
			Access: bonsai.ClusterAccess{
				Host:   "third-testing-clust-1234567890.us-east-1.bonsaisearch.net",
				Port:   443,
				Scheme: "https",
			},
			State: bonsai.ClusterStateProvisioned,
		},
	}
	clusters, err := s.client.Cluster.All(context.Background())
	s.NoError(err, "successfully get all clusters")
	s.Len(clusters, 3)

	s.ElementsMatch(expect, clusters, "elements in expect match elements in received clusters")

	// Comparisons on the individual struct level are much easier to debug
	for i, cluster := range clusters {
		s.Run(fmt.Sprintf("Cluster #%d", i), func() {
			s.Equal(expect[i], cluster)
		})
	}
}

func (s *ClientMockTestSuite) TestClusterClient_GetBySlug() {
	const targetClusterSlug = "second-testing-clust-1234567890"

	urlPath, err := url.JoinPath(bonsai.ClusterAPIBasePath, targetClusterSlug)
	s.NoError(err, "successfully resolved path")

	s.serveMux.Get(urlPath, func(w http.ResponseWriter, _ *http.Request) {
		respStr := fmt.Sprintf(`
				{			
					"slug": "%s",			
					"name": "second_testing_cluster",			
					"uri": "https://api.bonsai.io/clusters/second-testing-clust-1234567890",			
					"plan": {		 	 
						"slug": "sandbox-aws-us-east-1",			  
						"uri": "https://api.bonsai.io/plans/sandbox-aws-us-east-1"			
					},	 	 
					"release": {			  
						"version": "7.2.0",			  
						"slug": "elasticsearch-7.2.0",	
						"package_name": "7.2.0",	
						"service_type": "elasticsearch",			  
						"uri": "https://api.bonsai.io/releases/elasticsearch-7.2.0"			
					},  		
					"space": {			  
						"path": "omc/bonsai/us-east-1/common",			  
						"region": "aws-us-east-1",			  
						"uri": "https://api.bonsai.io/spaces/omc/bonsai/us-east-1/common"			
					},	  	
					"stats": {			  
						"docs": 0,			  
						"shards_used": 0,			  
						"data_bytes_used": 0			
						},	    
					"access": {			  
						"host": "second-testing-clust-1234567890.us-east-1.bonsaisearch.net",			  
						"port": 443,			  
						"scheme": "https"			
					},	  	
					"state": "PROVISIONED"		
				}
		`, targetClusterSlug)

		resp := &bonsai.Cluster{}
		err = json.Unmarshal([]byte(respStr), resp)
		s.NoError(err, "unmarshals json into bonsai.Space")

		err = json.NewEncoder(w).Encode(resp)
		s.NoError(err, "encodes bonsai.Space into json on the writer")
	})

	expect := bonsai.Cluster{
		Slug: "second-testing-clust-1234567890",
		Name: "second_testing_cluster",
		URI:  "https://api.bonsai.io/clusters/second-testing-clust-1234567890",
		Plan: bonsai.Plan{
			Slug:              "sandbox-aws-us-east-1",
			AvailableReleases: []bonsai.Release{},
			AvailableSpaces:   []bonsai.Space{},
			URI:               "https://api.bonsai.io/plans/sandbox-aws-us-east-1",
		},
		Release: bonsai.Release{
			Version:     "7.2.0",
			Slug:        "elasticsearch-7.2.0",
			PackageName: "7.2.0",
			ServiceType: "elasticsearch",
			URI:         "https://api.bonsai.io/releases/elasticsearch-7.2.0",
		},
		Space: bonsai.Space{
			Path:   "omc/bonsai/us-east-1/common",
			Region: "aws-us-east-1",
			URI:    "https://api.bonsai.io/spaces/omc/bonsai/us-east-1/common",
		},
		Stats: bonsai.ClusterStats{
			Docs:          0,
			ShardsUsed:    0,
			DataBytesUsed: 0,
		},
		Access: bonsai.ClusterAccess{
			Host:   "second-testing-clust-1234567890.us-east-1.bonsaisearch.net",
			Port:   443,
			Scheme: "https",
		},
		State: bonsai.ClusterStateProvisioned,
	}

	resultResp, err := s.client.Cluster.GetBySlug(context.Background(), targetClusterSlug)
	s.NoError(err, "successfully get cluster by path")

	s.Equal(expect, resultResp, "elements in expect match elements in received cluster response")
}

func (s *ClientMockTestSuite) TestClusterClient_Create() {
	s.serveMux.Post(bonsai.ClusterAPIBasePath, func(w http.ResponseWriter, _ *http.Request) {
		respStr := `
			{
				"message": "Your cluster is being provisioned.",
				"monitor": "https://api.bonsai.io/clusters/test-5-x-3968320296",
				"access": {
					"user": "utji08pwu6",
					"pass": "18v1fbey2y",
					"host": "test-5-x-3968320296",
					"port": 443,
					"scheme": "https",
					"url": "https://utji08pwu6:18v1fbey2y@test-5-x-3968320296.us-east-1.bonsaisearch.net:443"
				},
				"status": 202
			}
		`

		resp := &bonsai.ClustersResultCreate{}
		err := json.Unmarshal([]byte(respStr), resp)
		s.NoError(err, "unmarshals json into bonsai.ClustersResultCreate")

		err = json.NewEncoder(w).Encode(resp)
		s.NoError(err, "encodes bonsai.ClustersResultCreate into json on the writer")
	})

	expect := bonsai.ClustersResultCreate{
		Message: "Your cluster is being provisioned.",
		Monitor: "https://api.bonsai.io/clusters/test-5-x-3968320296",
		Access: bonsai.ClusterAccess{
			Host:     "test-5-x-3968320296",
			Port:     443,
			Scheme:   "https",
			Username: "utji08pwu6",
			Password: "18v1fbey2y",
			URL:      "https://utji08pwu6:18v1fbey2y@test-5-x-3968320296.us-east-1.bonsaisearch.net:443",
		},
	}

	resultResp, err := s.client.Cluster.Create(context.Background(), bonsai.ClusterCreateOpts{
		Name:    "test-5-x-3968320296",
		Plan:    "sandbox-aws-us-east-1",
		Space:   "omc/bonsai/us-east-1/common",
		Release: "elasticsearch-7.2.0",
	})
	s.NoError(err, "successfully execute create cluster request")

	s.Equal(expect, resultResp, "elements in expect match elements in received cluster create response")
}

func (s *ClientMockTestSuite) TestClusterClient_Update() {
	const targetClusterSlug = "second-testing-clust-1234567890"

	urlPath, err := url.JoinPath(bonsai.ClusterAPIBasePath, targetClusterSlug)
	s.NoError(err, "successfully resolved path")

	s.serveMux.Put(urlPath, func(w http.ResponseWriter, _ *http.Request) {
		respStr := `
		{
			"message": "Your cluster is being updated.",
			"monitor": "https://api.bonsai.io/clusters/test-5-x-3968320296",
			"status": 202
		}
		`

		resp := &bonsai.ClustersResultUpdate{}
		err = json.Unmarshal([]byte(respStr), resp)
		s.NoError(err, "unmarshals json into bonsai.ClustersResultUpdate")

		err = json.NewEncoder(w).Encode(resp)
		s.NoError(err, "encodes bonsai.ClustersResultUpdate into json on the writer")
	})

	expect := bonsai.ClustersResultUpdate{
		Message: "Your cluster is being updated.",
		Monitor: "https://api.bonsai.io/clusters/test-5-x-3968320296",
	}

	resultResp, err := s.client.Cluster.Update(context.Background(), targetClusterSlug, bonsai.ClusterUpdateOpts{
		Name: "test-5-x-3968320296",
		Plan: "sandbox-aws-us-east-2",
	})
	s.NoError(err, "successfully execute create cluster request")

	s.Equal(expect, resultResp, "items in expect match items in received cluster update response")
}

func (s *ClientMockTestSuite) TestClusterClient_Delete() {
	const targetClusterSlug = "second-testing-clust-1234567890"

	reqPath, err := url.JoinPath(bonsai.ClusterAPIBasePath, targetClusterSlug)
	s.NoError(err, "successfully resolved path")

	s.serveMux.Delete(reqPath, func(w http.ResponseWriter, _ *http.Request) {
		respStr := fmt.Sprintf(`
		{
			"message": "Your cluster is being deprovisioned.",
			"monitor": "%s",
			"status": 202
		}
		`, targetClusterSlug)

		resp := &bonsai.ClustersResultDestroy{}
		err = json.Unmarshal([]byte(respStr), resp)
		s.NoError(err, "unmarshals json into bonsai.ClustersResultDestroy")

		err = json.NewEncoder(w).Encode(resp)
		s.NoError(err, "encodes bonsai.ClustersResultDestroy into json on the writer")
	})

	expect := bonsai.ClustersResultDestroy{
		Message: "Your cluster is being deprovisioned.",
		Monitor: targetClusterSlug,
	}

	resultResp, err := s.client.Cluster.Destroy(context.Background(), targetClusterSlug)
	s.NoError(err, "successfully execute create cluster request")

	s.Equal(expect, resultResp, "items in expect match items in received cluster update response")
}

// VCR Tests.
func (s *ClientVCRTestSuite) TestClusterClient_All() {
	ctx := context.Background()

	plans, err := s.client.Cluster.All(ctx)
	s.NoError(err, "successfully get all clusters")
	assertGolden(s, plans)
}

func (s *ClientVCRTestSuite) TestClusterClient_GetBySlug() {
	ctx := context.Background()

	plan, err := s.client.Cluster.GetBySlug(ctx, "dcek-group-llc-5240651189")
	s.NoError(err, "successfully get cluster")
	assertGolden(s, plan)
}

func (s *ClientVCRTestSuite) TestClusterClient_Create() {
	ctx := context.Background()

	plan, err := s.client.Cluster.Create(ctx, bonsai.ClusterCreateOpts{
		Name:    "bonsai-api-go-test-cluster",
		Plan:    "standard-nano-comped",
		Space:   "omc/bonsai/us-east-1/common",
		Release: "opensearch-2.6.0-mt",
	})
	s.NoError(err, "successfully get cluster")
	assertGolden(s, plan)
}

func (s *ClientVCRTestSuite) TestClusterClient_Update() {
	ctx := context.Background()

	plan, err := s.client.Cluster.Update(ctx, "bonsai-api-go-9994392953", bonsai.ClusterUpdateOpts{
		Name: "bonsai-api-go-test-cluster-updated",
		Plan: "standard-nano-comped",
	})
	s.NoError(err, "successfully get cluster")
	assertGolden(s, plan)
}

func (s *ClientVCRTestSuite) TestClusterClient_Delete() {
	ctx := context.Background()

	plan, err := s.client.Cluster.Destroy(ctx, "bonsai-api-go-9994392953")
	s.NoError(err, "successfully get cluster")
	assertGolden(s, plan)
}
