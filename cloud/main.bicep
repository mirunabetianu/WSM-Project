targetScope = 'managementGroup'

// params
param subscriptionID string
param resourceGroupName string

@description('The location of this resource group')
param location string

@description('Object containing boolean fields indicating whether the corresponding resource should be deployed')
param featureFlag object = {
  eventGrid: false
  aks: false
  serviceBus: false
}

@description('The name for the aks.')
param aksName string = 'aks-kubernetes-cluster'

// vars
var skuTier = 'Basic'

// templates
// service bus module
module serviceBusModule 'templates/service_bus.bicep' = if (featureFlag.serviceBus) {
  name: 'serviceBusDeployment'
  params: {
    location: location
    skuName: skuTier
    skuTier: skuTier
  }
  scope: resourceGroup(subscriptionID, resourceGroupName)
}

// azure kubernetes service module
module aksModule 'templates/aks.bicep' = if (featureFlag.aks) {
  name: 'aksDeployment'
  params: {
    name: aksName
    location: location
  }
  scope: resourceGroup(subscriptionID, resourceGroupName)
}

// event grid module
module eventGridModule 'templates/event_grid.bicep' = if (featureFlag.eventGrid) {
  name: 'eventGridDeployment'
  params: {
    location: location
    source: aksModule.outputs.aksInfo.id
    topicType: 'Microsoft.EventGrid/eventSubscriptions'
    subscriptions: []
  }
  scope: resourceGroup(subscriptionID, resourceGroupName)
}
