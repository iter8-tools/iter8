# Notifiers
Iter8 supports sending notifications to external applications. This documentation describes how you can setup the notifier and gets status updates from the iter8 controller.

## Step 1: Setup External Webhook
Now only Slack integration is supported.
### Setup Slack Webhook
1. Go to here[https://api.slack.com/apps] to create a new Slack app.  
2. Choose the Slack Workspace you would like to register the app.   
3. After creating the app, click into `Incoming Webhooks` under the `Add features and functionality` section.  
4. Toggle the switch on the top right corner to `on` to activate the incoming webhook ability.   
5. Go to the button of the page and click the `Add New Webhook to Workspace` botton. Here you will need to choose the channel to recieve the notifications pushed by the app. By clicking `Allow`, this channel will be granted the right.  
6. Now you should see the webhook you just created shown in the page. We will use the Webhook URL to setup the notifier in the controller.

## Step 2: Setup Notifiers 
The iter8 controller reads notifier settings from the configmap `iter8config-notifiers` under namespace `iter8`. This configmap is deployed along with installation of iter8 controller. You can add/update/delete the notifier settings while the controller is running. The changes will be reflected immediately.

Let's take a look at an example.
```yaml
apiVersion: v1
 kind: ConfigMap
 metadata:
   name: iter8config-notifiers
   namespace: iter8
 data:
  # Name of the channel can be any unique value;
  # This is used to distinguished this channel setting from others.
  name_of_channel: |-
    # The name of the external application.
    # required; Now only slack is supoorted.
    notifier: slack
    # The Webhook url of the application.
    # required;
    url: https://hooks.slack.com/services/TXXXXX/BXXXXXX/xxxxxxxx
    # Level of notification frequency.
    # optional;
    # options are: error, warning, normal, verbose
    level: normal
    # Only experiments under this namespace can push notifications to this channel
    # optional; 
    namespace: bookinfo-iter8
    # Only experiments with these labels can push notifications to this channel
    # optional; 
    labels:
      foo: bar
```

The data part is a map, where the key is the name of the channel while the value is the configuration of the channel. The name can be any unique identifier. For the configuration, `notifier` and `url` are required fields. Others are optional. 

`namespace` and `labels` are filtering options for Experiment instances. Only experiments with all requirements satisfied can have status pushed to this channel. 

`level` specifies the severity level of information that the channel is going to digest. 4 levels are available: `error`, `warning`, `normal`, `verbose`. `normal` is the default option. There are circumstances that the controller is going to send notifications, and their relationships with the severity levels are shown as following:

Experiment Succeeded  
Experiment Failed  
Targets Not Found  
Sync Metrics Error  
Routing Rules Error  
Analytics Service Error   
------------------------------------- error  
Progress Failure  
------------------------------------- warning  
Progress Succeeded  
------------------------------------- normal  
Targets Found  
Analytics Service Running  
Iteration Update  
Sync Metrics Succeeded  
Routing Rules Ready  
------------------------------------- verbose  

When situations defined above the specified severity level happen, corresponding notifications will be pushed to the receiver end. 