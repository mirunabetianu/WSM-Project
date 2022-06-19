@description('The name of this resource.')
var name = 'sb-wsm-${uniqueString(resourceGroup().id)}'

@description('The location that this resource will be deployed in.')
param location string = resourceGroup().location

@allowed([
  'Basic'
  'Premium'
  'Standard'
])
param skuName string

@allowed([
  'Basic'
  'Premium'
  'Standard'
])
param skuTier string

resource serviceBus 'Microsoft.ServiceBus/namespaces@2021-11-01' = {
  name: name
  location: location
  tags: {}
  sku: {
    capacity: 1
    name: skuName
    tier: skuTier
  }
  // identity: {
  //   type: 'string'
  //   userAssignedIdentities: {}
  // }
  properties: {
    // alternateName: 'string'
    // disableLocalAuth: false
    // encryption: {
    //   keySource: 'Microsoft.KeyVault'
    //   keyVaultProperties: [
    //     {
    //       identity: {
    //         userAssignedIdentity: 'string'
    //       }
    //       keyName: 'string'
    //       keyVaultUri: 'string'
    //       keyVersion: 'string'
    //     }
    //   ]
    //   requireInfrastructureEncryption: bool
    // }
    // privateEndpointConnections: [
    //   {
    //     properties: {
    //       privateEndpoint: {
    //         id: 'string'
    //       }
    //       privateLinkServiceConnectionState: {
    //         description: 'string'
    //         status: 'string'
    //       }
    //       provisioningState: 'string'
    //     }
    //   }
    // ]
    zoneRedundant: false
  }
}
