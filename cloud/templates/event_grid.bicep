var topicName = 'evgst-wsm-dl-${uniqueString(resourceGroup().id)}'

@description('The location that this resource will be deployed in.')
param location string = resourceGroup().location

@description('The source of this event grid system topic.')
param source string

@description('The type of topic that should be created for this event grid system topic.')
param topicType string

@description('''
An array of objects containing information about the subscriptions that need to be created.
Subscription: {
    name: 'string',
    resourceId: 'string',
    includedEventTypes: [inlcuded events],
    containerName: 'string'
}
''')
param subscriptions array

resource eventGridSystemTopics 'Microsoft.EventGrid/systemTopics@2021-12-01' = {
  name: topicName
  location: location
  tags: {}
  properties: {
    source: source
    topicType: topicType
  }
  // identity: {
  // principalId: 'string'
  // tenantId: 'string'
  // type: 'string'
  // userAssignedIdentities: {}
  // }
}

resource eventSubscription 'Microsoft.EventGrid/systemTopics/eventSubscriptions@2021-12-01' = [for subscription in subscriptions: {
  name: subscription.name
  parent: eventGridSystemTopics
  properties: {
    destination: {
      endpointType: 'AzureFunction'
      // For remaining properties, see EventSubscriptionDestination objects
      properties: {
        resourceId: subscription.resourceId
        //   deliveryAttributeMappings: [
        //     {
        //       name: 'string'
        //       type: 'string'
        //       // For remaining properties, see DeliveryAttributeMapping objects
        //     }
        //   ]
        // maxEventsPerBatch: int
        // preferredBatchSizeInKilobytes: int
      }
    }
    eventDeliverySchema: 'EventGridSchema'
    filter: {
      includedEventTypes: subscription.includedEventTypes
      subjectBeginsWith: '/blobServices/default/containers/${subscription.containerName}/'
      // subjectEndsWith: 'string'
      // enableAdvancedFilteringOnArrays: bool
      // isSubjectCaseSensitive: bool
      // advancedFilters: [
      //   {
      //     key: 'string'
      //     operatorType: 'string'
      //     // For remaining properties, see AdvancedFilter objects
      //   }
      // ]
    }
  }
}]

output systemTopic object = {
  source: eventGridSystemTopics.properties.source
  name: eventGridSystemTopics.name
  id: eventGridSystemTopics.id
  topicType: eventGridSystemTopics.properties.topicType
}
