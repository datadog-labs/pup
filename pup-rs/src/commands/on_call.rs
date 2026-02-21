use anyhow::Result;
use datadog_api_client::datadogV2::api_teams::{
    GetTeamMembershipsOptionalParams, ListTeamsOptionalParams, TeamsAPI,
};
use datadog_api_client::datadogV2::model::{
    RelationshipToUserTeamUser, RelationshipToUserTeamUserData, TeamCreate, TeamCreateAttributes,
    TeamCreateRequest, TeamType, TeamUpdate, TeamUpdateAttributes, TeamUpdateRequest,
    UserTeamAttributes, UserTeamCreate, UserTeamRelationships, UserTeamRequest, UserTeamRole,
    UserTeamType, UserTeamUpdate, UserTeamUpdateRequest, UserTeamUserType,
};

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn teams_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TeamsAPI::with_client_and_config(dd_cfg, c),
        None => TeamsAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_teams(ListTeamsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list teams: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn teams_get(cfg: &Config, team_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TeamsAPI::with_client_and_config(dd_cfg, c),
        None => TeamsAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_team(team_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get team: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn teams_delete(cfg: &Config, team_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TeamsAPI::with_client_and_config(dd_cfg, c),
        None => TeamsAPI::with_config(dd_cfg),
    };
    api.delete_team(team_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete team: {e:?}"))?;
    println!("Team '{team_id}' deleted successfully.");
    Ok(())
}

pub async fn teams_create(cfg: &Config, name: &str, handle: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TeamsAPI::with_client_and_config(dd_cfg, c),
        None => TeamsAPI::with_config(dd_cfg),
    };
    let attrs = TeamCreateAttributes::new(handle.to_string(), name.to_string());
    let data = TeamCreate::new(attrs, TeamType::TEAM);
    let body = TeamCreateRequest::new(data);
    let resp = api
        .create_team(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create team: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn teams_update(cfg: &Config, team_id: &str, name: &str, handle: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TeamsAPI::with_client_and_config(dd_cfg, c),
        None => TeamsAPI::with_config(dd_cfg),
    };
    let attrs = TeamUpdateAttributes::new(handle.to_string(), name.to_string());
    let data = TeamUpdate::new(attrs, TeamType::TEAM);
    let body = TeamUpdateRequest::new(data);
    let resp = api
        .update_team(team_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update team: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn memberships_list(cfg: &Config, team_id: &str, page_size: i64) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TeamsAPI::with_client_and_config(dd_cfg, c),
        None => TeamsAPI::with_config(dd_cfg),
    };
    let params = GetTeamMembershipsOptionalParams::default().page_size(page_size);
    let resp = api
        .get_team_memberships(team_id.to_string(), params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list memberships: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn memberships_add(
    cfg: &Config,
    team_id: &str,
    user_id: &str,
    role: Option<String>,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TeamsAPI::with_client_and_config(dd_cfg, c),
        None => TeamsAPI::with_config(dd_cfg),
    };
    let mut attrs = UserTeamAttributes::new();
    if let Some(r) = role {
        let team_role = match r.to_lowercase().as_str() {
            "admin" => UserTeamRole::ADMIN,
            _ => UserTeamRole::ADMIN,
        };
        attrs = attrs.role(Some(team_role));
    }
    let user_data = RelationshipToUserTeamUserData::new(user_id.to_string(), UserTeamUserType::USERS);
    let user_rel = RelationshipToUserTeamUser::new(user_data);
    let relationships = UserTeamRelationships::new().user(user_rel);
    let data = UserTeamCreate::new(UserTeamType::TEAM_MEMBERSHIPS)
        .attributes(attrs)
        .relationships(relationships);
    let body = UserTeamRequest::new(data);
    let resp = api
        .create_team_membership(team_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to add membership: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn memberships_update(
    cfg: &Config,
    team_id: &str,
    user_id: &str,
    role: &str,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TeamsAPI::with_client_and_config(dd_cfg, c),
        None => TeamsAPI::with_config(dd_cfg),
    };
    let team_role = match role.to_lowercase().as_str() {
        "admin" => UserTeamRole::ADMIN,
        _ => UserTeamRole::ADMIN,
    };
    let attrs = UserTeamAttributes::new().role(Some(team_role));
    let data = UserTeamUpdate::new(UserTeamType::TEAM_MEMBERSHIPS).attributes(attrs);
    let body = UserTeamUpdateRequest::new(data);
    let resp = api
        .update_team_membership(team_id.to_string(), user_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update membership: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn memberships_remove(cfg: &Config, team_id: &str, user_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TeamsAPI::with_client_and_config(dd_cfg, c),
        None => TeamsAPI::with_config(dd_cfg),
    };
    api.delete_team_membership(team_id.to_string(), user_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to remove membership: {e:?}"))?;
    println!("Membership for user {user_id} removed from team {team_id}.");
    Ok(())
}
