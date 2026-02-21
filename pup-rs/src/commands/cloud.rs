use anyhow::Result;
use datadog_api_client::datadogV1::api_aws_integration::{AWSIntegrationAPI, ListAWSAccountsOptionalParams};
use datadog_api_client::datadogV1::api_gcp_integration::GCPIntegrationAPI;
use datadog_api_client::datadogV1::api_azure_integration::AzureIntegrationAPI;

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn aws_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => AWSIntegrationAPI::with_client_and_config(dd_cfg, c),
        None => AWSIntegrationAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_aws_accounts(ListAWSAccountsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list AWS accounts: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn gcp_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => GCPIntegrationAPI::with_client_and_config(dd_cfg, c),
        None => GCPIntegrationAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_gcp_integration()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list GCP integrations: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn azure_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => AzureIntegrationAPI::with_client_and_config(dd_cfg, c),
        None => AzureIntegrationAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_azure_integration()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list Azure integrations: {e:?}"))?;
    formatter::print_json(&resp)
}
