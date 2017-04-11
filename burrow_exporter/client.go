package burrow_exporter

import (
	"net/http"
	"net/url"
	"time"

	"path"

	"encoding/json"
	"errors"

	"fmt"

	log "github.com/Sirupsen/logrus"
)

/*
Request	Method	URL Format
Healthcheck	GET	/burrow/admin
List ClustersResp	GET	/v2/kafka
Kafka Cluster Detail	GET	/v2/kafka/(cluster)
List Consumers	GET	/v2/kafka/(cluster)/consumer
Remove Consumer Group	DELETE	/v2/kafka/(cluster)/consumer/(group)
List Consumer Topics	GET	/v2/kafka/(cluster)/consumer/(group)/topic
Consumer Topic Detail	GET	/v2/kafka/(cluster)/consumer/(group)/topic/(topic)
Consumer Group Status	GET	/v2/kafka/(cluster)/consumer/(group)/status /v2/kafka/(cluster)/consumer/(group)/lag
List Cluster Topics	GET	/v2/kafka/(cluster)/topic
Cluster Topic Detail	GET	/v2/kafka/(cluster)/topic/(topic)
List ClustersResp	GET
*/

type BurrowResp struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type ClustersResp struct {
	BurrowResp
	Clusters []string `json:"clusters"`
}

type ClusterDetails struct {
	Brokers       []string `json:"brokers"`
	Zookeepers    []string `json:"zookeepers"`
	BrokerPort    int      `json:"broker_port"`
	ZookeeperPort int      `json:"zookeeper_port"`
	OffsetsTopic  string   `json:"offsets_topic"`
}

type ClusterDetailsResp struct {
	BurrowResp
	Cluster ClusterDetails `json:"cluster"`
}

type ConsumerGroupsResp struct {
	BurrowResp
	ConsumerGroups []string `json:"consumers"`
}

type ConsumerGroupTopicsResp struct {
	BurrowResp
	Topics []string `json:"topics"`
}

type ConsumerGroupTopicDetailsResp struct {
	BurrowResp
	Offsets []int64 `json:"offsets"`
}

type Offset struct {
	Offset    int64 `json:"offset"`
	Timestamp int64 `json:"timestamp"`
	Lag       int64 `json:"lag"`
}

type ConsumerGroupStatus struct {
	Cluster    string      `json:"cluster"`
	Group      string      `json:"group"`
	Status     string      `json:"status"`
	Complete   bool        `json:"complete"`
	MaxLag     Partition   `json:"maxlag"`
	Partitions []Partition `json:"partitions"`
	TotalLag   int64       `json:"total_lag"`
}

type Partition struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Status    string `json:"status"`
	Start     Offset `json:"start"`
	End       Offset `json:"end"`
}

type ConsumerGroupStatusResp struct {
	BurrowResp
	Status ConsumerGroupStatus `json:"status"`
}

type ClusterTopicDetailsResp struct {
	BurrowResp
	Offsets []int64 `json:"offsets"`
}

type BurrowClient struct {
	baseUrl string
	client  *http.Client
}

func (bc *BurrowClient) buildUrl(endpoint string) (string, error) {
	parsedUrl, err := url.Parse(bc.baseUrl)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"baseUrl": bc.baseUrl,
		}).Error("error parsing base url")
		return "", err
	}

	parsedUrl.Path = path.Join(parsedUrl.Path, endpoint)

	return parsedUrl.String(), nil
}

func (bc *BurrowClient) getJsonReq(endpoint string, dest interface{}) error {
	resp, err := bc.client.Get(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"endpoint": endpoint,
		}).Error("error making request")
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(dest)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("error decoding json")
		return err
	}

	return nil
}

func (bc *BurrowClient) HealthCheck() (bool, error) {
	endpoint, err := bc.buildUrl("/burrow/admin")
	if err != nil {
		return false, err
	}

	_, err = bc.client.Get(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"endpoint": endpoint,
		})
		return false, err
	}

	return true, nil
}

func (bc *BurrowClient) ListClusters() (*ClustersResp, error) {
	endpoint, err := bc.buildUrl("/v2/kafka")
	if err != nil {
		return nil, err
	}

	clusters := &ClustersResp{}
	err = bc.getJsonReq(endpoint, clusters)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("error retrieving cluster details")
		return nil, err
	}

	if clusters.Error {
		log.WithFields(log.Fields{
			"err": clusters.Message,
		}).Error("error retrieving clusters")
		return nil, errors.New(clusters.Message)
	}

	return clusters, nil
}

func (bc *BurrowClient) ClusterDetails(cluster string) (*ClusterDetailsResp, error) {
	endpoint, err := bc.buildUrl(fmt.Sprintf("/v2/kafka/%s", cluster))
	if err != nil {
		return nil, err
	}

	clusterDetails := &ClusterDetailsResp{}
	err = bc.getJsonReq(endpoint, clusterDetails)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"cluster": cluster,
		}).Error("error retrieving cluster details")
		return nil, err
	}

	if clusterDetails.Error {
		log.WithFields(log.Fields{
			"err":     clusterDetails.Message,
			"cluster": cluster,
		}).Error("error retrieving cluster details")
		return nil, errors.New(clusterDetails.Message)
	}

	return clusterDetails, nil
}

func (bc *BurrowClient) ListConsumers(cluster string) (*ConsumerGroupsResp, error) {
	endpoint, err := bc.buildUrl(fmt.Sprintf("/v2/kafka/%s/consumer", cluster))
	if err != nil {
		return nil, err
	}

	consumers := &ConsumerGroupsResp{}
	err = bc.getJsonReq(endpoint, consumers)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"cluster": cluster,
		}).Error("error retrieving consumer groups")
		return nil, err
	}

	if consumers.Error {
		log.WithFields(log.Fields{
			"err":     consumers.Message,
			"cluster": cluster,
		}).Error("error retrieving cluster consumer groups")
		return nil, errors.New(consumers.Message)
	}

	return consumers, nil
}

func (bc *BurrowClient) ListConsumerTopics(cluster, consumerGroup string) (*ConsumerGroupTopicsResp, error) {
	endpoint, err := bc.buildUrl(fmt.Sprintf("/v2/kafka/%s/consumer/%s/topic", cluster, consumerGroup))
	if err != nil {
		return nil, err
	}

	consumerTopics := &ConsumerGroupTopicsResp{}
	err = bc.getJsonReq(endpoint, consumerTopics)
	if err != nil {
		log.WithFields(log.Fields{
			"err":           err,
			"cluster":       cluster,
			"consumerGroup": consumerGroup,
		}).Error("error retrieving consumer group topics")
		return nil, err
	}

	if consumerTopics.Error {
		log.WithFields(log.Fields{
			"err":           consumerTopics.Message,
			"consumerGroup": consumerGroup,
			"cluster":       cluster,
		}).Error("error retriving consumer group topics")
		return nil, errors.New(consumerTopics.Message)
	}

	return consumerTopics, nil
}

func (bc *BurrowClient) ConsumerGroupTopicDetails(cluster, consumerGroup, topic string) (*ConsumerGroupTopicDetailsResp, error) {
	endpoint, err := bc.buildUrl(fmt.Sprintf("/v2/kafka/%s/consumer/%s/topic/%s", cluster, consumerGroup, topic))
	if err != nil {
		return nil, err
	}

	topicDetails := &ConsumerGroupTopicDetailsResp{}
	err = bc.getJsonReq(endpoint, topicDetails)
	if err != nil {
		log.WithFields(log.Fields{
			"err":           err,
			"cluster":       cluster,
			"consumerGroup": consumerGroup,
			"topic":         topic,
		}).Error("error retrieving consumer group topic details")
		return nil, err
	}

	if topicDetails.Error {
		log.WithFields(log.Fields{
			"err":           err,
			"cluster":       cluster,
			"consumerGroup": consumerGroup,
			"topic":         topic,
		}).Error("error retrieving consumer group topic details")
		return nil, errors.New(topicDetails.Message)
	}

	return topicDetails, nil
}

func (bc *BurrowClient) ConsumerGroupStatus(cluster, consumerGroup string) (*ConsumerGroupStatusResp, error) {
	endpoint, err := bc.buildUrl(fmt.Sprintf("/v2/kafka/%s/consumer/%s/status", cluster, consumerGroup))
	if err != nil {
		return nil, err
	}

	status := &ConsumerGroupStatusResp{}
	err = bc.getJsonReq(endpoint, status)
	if err != nil {
		log.WithFields(log.Fields{
			"err":           err,
			"cluster":       cluster,
			"consumerGroup": consumerGroup,
		}).Error("error retrieving consumer group status")
		return nil, err
	}

	if status.Error {
		log.WithFields(log.Fields{
			"err":           err,
			"cluster":       cluster,
			"consumerGroup": consumerGroup,
		}).Error("error retrieving consumer group status")
		return nil, errors.New(status.Message)
	}

	return status, nil
}

func (bc *BurrowClient) ConsumerGroupLag(cluster, consumerGroup string) (*ConsumerGroupStatusResp, error) {
	endpoint, err := bc.buildUrl(fmt.Sprintf("/v2/kafka/%s/consumer/%s/lag", cluster, consumerGroup))
	if err != nil {
		return nil, err
	}

	status := &ConsumerGroupStatusResp{}
	err = bc.getJsonReq(endpoint, status)
	if err != nil {
		log.WithFields(log.Fields{
			"err":           err,
			"cluster":       cluster,
			"consumerGroup": consumerGroup,
		}).Error("error retrieving consumer group status")
		return nil, err
	}

	if status.Error {
		log.WithFields(log.Fields{
			"err":           err,
			"cluster":       cluster,
			"consumerGroup": consumerGroup,
		}).Error("error retrieving consumer group status")
		return nil, errors.New(status.Message)
	}

	return status, nil
}

func (bc *BurrowClient) ClusterTopicDetails(cluster, topic string) (*ClusterTopicDetailsResp, error) {
	endpoint, err := bc.buildUrl(fmt.Sprintf("/v2/kafka/%s/topic/%s", cluster, topic))
	if err != nil {
		return nil, err
	}

	topicDetails := &ClusterTopicDetailsResp{}
	err = bc.getJsonReq(endpoint, topicDetails)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"cluster": cluster,
			"topic":   topic,
		}).Error("error retrieving consumer group topic details")
		return nil, err
	}

	if topicDetails.Error {
		log.WithFields(log.Fields{
			"err":     err,
			"cluster": cluster,
			"topic":   topic,
		}).Error("error retrieving consumer group topicDetails")
		return nil, errors.New(topicDetails.Message)
	}

	return topicDetails, nil
}

func MakeBurrowClient(baseUrl string) *BurrowClient {
	return &BurrowClient{
		baseUrl: baseUrl,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
