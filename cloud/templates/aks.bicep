@description('The location that this aks will be deployed in.')
param location string = resourceGroup().location

@description('The name for this AKS.')
param name string = 'aks-kubernetes-cluster'

resource aks 'Microsoft.ContainerService/managedClusters@2022-03-02-preview' = {
  name: name
  location: location
  tags: {}
  sku: {
    name: 'Basic'
    tier: 'Free'
  }
  identity: {
    type: 'None'
    userAssignedIdentities: {}
  }
  properties: {
    kubernetesVersion: '1.22.6'
    publicNetworkAccess: 'Enabled'
    servicePrincipalProfile: {
      clientId: 'string (required)'
      secret: 'string'
    }
  }
}

output aksInfo object = {
  id: aks.id
  name: aks.name
}
