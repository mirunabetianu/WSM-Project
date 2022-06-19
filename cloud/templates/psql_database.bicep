@description('Server Name for Azure database for PostgreSQL')
param serverName string = 'psql-main-database-001'

@description('Database administrator login name')
@minLength(1)
param administratorLogin string = 'group7_wsm'

@description('Database administrator password')
@minLength(8)
@secure()
param administratorLoginPassword string

@description('Azure database for PostgreSQL compute capacity in vCores (1,2,4,8,16,32)')
param skuCapacity int = 1

@description('Azure database for PostgreSQL sku name ')
param skuName string = 'Standard_B1ms'

@description('Azure database for PostgreSQL Sku Size ')
param skuSizeMB int = 51200

@description('Azure database for PostgreSQL pricing tier')
@allowed([
  'Basic'
  'GeneralPurpose'
  'MemoryOptimized'
  'Burstable'
])
param skuTier string = 'Burstable'

@description('PostgreSQL version')
@allowed([
  '11'
  '12'
  '13'
])
param postgresqlVersion string = '13'

@description('Location for all resources.')
param location string = resourceGroup().location

@description('PostgreSQL Server backup retention days')
param backupRetentionDays int = 7

@description('Geo-Redundant Backup setting')
param geoRedundantBackup string = 'Disabled'

resource server 'Microsoft.DBforPostgreSQL/servers@2017-12-01' = {
  name: serverName
  location: location
  sku: {
    name: skuName
    tier: skuTier
    capacity: skuCapacity
  }
  properties: {
    createMode: 'Default'
    version: postgresqlVersion
    administratorLogin: administratorLogin
    administratorLoginPassword: administratorLoginPassword
    storageProfile: {
      storageMB: skuSizeMB
      backupRetentionDays: backupRetentionDays
      geoRedundantBackup: geoRedundantBackup
    }
  }
}
