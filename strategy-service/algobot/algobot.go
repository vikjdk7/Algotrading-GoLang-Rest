package algobot

import (
	"context"
	"fmt"
	"sync"

	"github.com/vikjdk7/Algotrading-GoLang-Rest/strategy-service/models"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

/*
var (
	kubeconfigpath = "./.kube/config"
)
*/

func StartBot(strategy models.Strategy, dealId string, asset string, api_key string, api_secret string, alpaca_url string, wg *sync.WaitGroup) string {
	fmt.Println(fmt.Sprintf("Starting bot for Deal Id: %s, Strategy: %s and exchange: %s with parameters:", dealId, strategy.StrategyName, strategy.SelectedExchangeName))

	clientset, errMsg := connectToK8s()
	if errMsg != "" {
		wg.Done()
		return errMsg
	}
	containerimage := "hedgina/algobot-job:latest"
	jobname := fmt.Sprintf("algobot-job-dealid-%s", dealId)
	errMsg = launchK8sJob(clientset, &containerimage, &jobname, strategy, dealId, asset, api_key, api_secret, alpaca_url)
	if errMsg != "" {
		wg.Done()
		return errMsg
	}
	wg.Done()
	return ""
}

func StopBot(dealId string) string {
	fmt.Println(fmt.Sprintf("Stopping bot and deleting Job for Deal Id: %s", dealId))
	//ChangeDealStatus(dealId, "cancelled")
	clientset, errMsg := connectToK8s()
	if errMsg != "" {
		return errMsg
	}
	jobname := fmt.Sprintf("algobot-job-dealid-%s", dealId)
	fg := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{PropagationPolicy: &fg}
	if err := clientset.BatchV1().Jobs("hedgina").Delete(context.TODO(), jobname, deleteOptions); err != nil {
		return err.Error()
	}
	return ""
}

func connectToK8s() (*kubernetes.Clientset, string) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	//config, err := clientcmd.BuildConfigFromFlags("", kubeconfigpath)
	if err != nil {
		return nil, "Failed to create K8s config"
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, "Failed to create K8s clientset"
	}
	return clientset, ""
}

func launchK8sJob(clientset *kubernetes.Clientset, image *string, jobname *string, strategy models.Strategy, dealId string, asset string, api_key string, api_secret string, alpaca_url string) string {
	jobs := clientset.BatchV1().Jobs("hedgina")
	var backOffLimit int32 = 0
	var ttlSecondsAfterFinished int32 = 259200 //3 Days = 259200 seconds. Job will be deleted 3 days after its completion
	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *jobname,
			Namespace: "hedgina",
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: *jobname + "-",
					Namespace:    "hedgina",
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "algobot-job",
							Image:           *image,
							ImagePullPolicy: v1.PullPolicy(v1.PullAlways),
							Env: []v1.EnvVar{
								{
									Name:  "base_order_size",
									Value: fmt.Sprintf("%v", strategy.BaseOrderSize),
								},
								{
									Name:  "target_profit_percent",
									Value: strategy.TargetProfit,
								},
								{
									Name:  "safety_order_size",
									Value: fmt.Sprintf("%v", strategy.SafetyOrderSize),
								},
								{
									Name:  "max_safety_order_count",
									Value: fmt.Sprintf("%v", strategy.MaxSafetyTradeCount),
								},
								{
									Name:  "max_active_safety_order_count",
									Value: fmt.Sprintf("%v", strategy.MaxActiveSafetyTradeCount),
								},
								{
									Name:  "price_deviation",
									Value: strategy.PriceDevation,
								},
								{
									Name:  "safety_order_step_scale",
									Value: fmt.Sprintf("%v", strategy.SafetyOrderStepScale),
								},
								{
									Name:  "safety_order_volume_scale",
									Value: fmt.Sprintf("%v", strategy.SafetyOrderVolumeScale),
								},
								{
									Name:  "stop_loss_percent",
									Value: strategy.StopLossPercent,
								},
								{
									Name:  "asset",
									Value: asset,
								},
								{
									Name:  "user_id",
									Value: strategy.UserId,
								},
								{
									Name:  "exchange_id",
									Value: strategy.SelectedExchange,
								},
								{
									Name:  "deal_id",
									Value: dealId,
								},
								{
									Name:  "strategy_id",
									Value: strategy.Id.Hex(),
								},
								{
									Name:  "strategy_name",
									Value: strategy.StrategyName,
								},
								{
									Name:  "alpaca_api_key",
									Value: api_key,
								},
								{
									Name:  "alpaca_api_secret",
									Value: api_secret,
								},
								{
									Name:  "alpaca_url",
									Value: alpaca_url,
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyOnFailure,
				},
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		},
	}
	_, err := jobs.Create(context.TODO(), jobSpec, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return "Failed to create K8s job."
	}
	fmt.Printf("Created K8S Job %s Successfully", *jobname)
	return ""
}
