use std::time::Duration;

use anyhow::{bail, Context, Result};
use clap::Parser;
use reqwest::header::{AUTHORIZATION, CONTENT_TYPE};
use serde::Deserialize;
use serde_json::{json, Value};

const DEFAULT_GRAPHQL_URL: &str = "http://127.0.0.1:8080/graphql";
const DEFAULT_TIMEOUT_MS: u64 = 10_000;
const PASSWORD: &str = "password123";
const CREATOR_DECKLIST: &str = "100, Island";
const JOINER_DECKLIST: &str = "100, Swamp";

const SIGNUP_MUTATION: &str = r#"
  mutation Signup($username: String!, $password: String!) {
    signup(username: $username, password: $password) {
      ID
      Username
      Token
    }
  }
"#;

const CREATE_GAME_MUTATION: &str = r#"
  mutation CreateGame($input: InputCreateGame!) {
    createGame(input: $input) {
      ID
      Players {
        ID
        Username
      }
      Turn {
        Player
        Phase
        Number
        Priority
      }
    }
  }
"#;

const JOIN_GAME_MUTATION: &str = r#"
  mutation JoinGame($input: InputJoinGame) {
    joinGame(input: $input) {
      ID
      Players {
        ID
        Username
      }
    }
  }
"#;

const GUEST_SESSION_MUTATION: &str = r#"
  mutation GuestSession($displayName: String, $sessionID: String!) {
    guestSession(displayName: $displayName, sessionID: $sessionID) {
      ID
      Username
      Token
    }
  }
"#;

const GAME_INVITE_QUERY: &str = r#"
  query GameInvite($gameID: String!, $sessionID: String!) {
    gameInvite(gameID: $gameID, sessionID: $sessionID) {
      ID
      Status
      PlayerCount
      Capacity
      PlayerDisplayNames
    }
  }
"#;

#[derive(Parser, Debug)]
#[command(name = "vedh-smoke", about = "vEDH GraphQL smoke runner")]
struct Args {
    #[arg(long, env = "VEDH_GRAPHQL_URL", default_value = DEFAULT_GRAPHQL_URL)]
    graphql_url: String,
    #[arg(long, env = "VEDH_SMOKE_TIMEOUT_MS", default_value_t = DEFAULT_TIMEOUT_MS)]
    timeout_ms: u64,
}

#[derive(Debug, Deserialize)]
struct GraphQlError {
    message: Option<String>,
}

#[derive(Debug, Deserialize)]
struct GraphQlResponse<T> {
    data: Option<T>,
    errors: Option<Vec<GraphQlError>>,
}

#[derive(Debug, Deserialize)]
struct SignupEnvelope {
    signup: SignupPayload,
}

#[derive(Debug, Deserialize)]
struct SignupPayload {
    #[serde(rename = "ID")]
    id: String,
    #[serde(rename = "Username")]
    username: String,
    #[serde(rename = "Token")]
    token: String,
}

#[derive(Debug, Deserialize)]
struct CreateGameEnvelope {
    #[serde(rename = "createGame")]
    create_game: GamePayload,
}

#[derive(Debug, Deserialize)]
struct JoinGameEnvelope {
    #[serde(rename = "joinGame")]
    join_game: GamePayload,
}

#[derive(Debug, Deserialize)]
struct GuestSessionEnvelope {
    #[serde(rename = "guestSession")]
    guest_session: SignupPayload,
}

#[derive(Debug, Deserialize)]
struct GameInviteEnvelope {
    #[serde(rename = "gameInvite")]
    game_invite: Option<GameInvitePayload>,
}

#[derive(Debug, Deserialize)]
struct GameInvitePayload {
    #[serde(rename = "ID")]
    id: String,
    #[serde(rename = "Status")]
    status: String,
    #[serde(rename = "PlayerCount")]
    player_count: i64,
    #[serde(rename = "Capacity")]
    capacity: i64,
    #[serde(rename = "PlayerDisplayNames")]
    player_display_names: Vec<String>,
}

#[derive(Debug, Deserialize)]
struct GamePayload {
    #[serde(rename = "ID")]
    id: String,
    #[serde(rename = "Players")]
    players: Vec<PlayerSummary>,
}

#[derive(Debug, Deserialize)]
struct PlayerSummary {
    #[serde(rename = "ID")]
    id: String,
    #[serde(rename = "Username")]
    username: String,
}

struct SmokeClient {
    http: reqwest::Client,
    graphql_url: String,
}

impl SmokeClient {
    fn new(graphql_url: String, timeout_ms: u64) -> Result<Self> {
        let http = reqwest::Client::builder()
            .timeout(Duration::from_millis(timeout_ms))
            .build()
            .context("failed to build HTTP client")?;
        Ok(Self { http, graphql_url })
    }

    async fn post_graphql<T: for<'de> Deserialize<'de>>(
        &self,
        query: &str,
        variables: Value,
        token: Option<&str>,
    ) -> Result<T> {
        let mut request = self
            .http
            .post(&self.graphql_url)
            .header(CONTENT_TYPE, "application/json")
            .json(&json!({ "query": query, "variables": variables }));

        if let Some(token) = token {
            request = request.header(AUTHORIZATION, format!("Bearer {token}"));
        }

        let response = request.send().await.map_err(|error| {
            anyhow::anyhow!("Connection error: {} | {}", error, self.graphql_url)
        })?;

        let status = response.status();
        let text = response
            .text()
            .await
            .context("failed to read GraphQL response")?;
        let payload: GraphQlResponse<T> = serde_json::from_str(&text).map_err(|_| {
            anyhow::anyhow!(
                "GraphQL response was not valid JSON ({}): {}",
                status,
                text.chars().take(200).collect::<String>()
            )
        })?;

        if !status.is_success() {
            bail!("GraphQL HTTP {}: {}", status, text);
        }

        if let Some(errors) = payload.errors {
            let joined = errors
                .into_iter()
                .map(|entry| {
                    entry
                        .message
                        .unwrap_or_else(|| "unknown GraphQL error".to_string())
                })
                .collect::<Vec<_>>()
                .join("; ");
            bail!("GraphQL error: {joined}");
        }

        payload.data.context("GraphQL response missing data")
    }
}

fn random_suffix() -> String {
    use std::time::{SystemTime, UNIX_EPOCH};
    let millis = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|duration| duration.as_millis())
        .unwrap_or_default();
    format!("{}_{}", millis, std::process::id())
}

async fn signup(client: &SmokeClient, username: &str) -> Result<SignupPayload> {
    let data: SignupEnvelope = client
        .post_graphql(
            SIGNUP_MUTATION,
            json!({ "username": username, "password": PASSWORD }),
            None,
        )
        .await?;

    if data.signup.id.is_empty() {
        bail!("signup did not return ID");
    }
    if data.signup.username != username {
        bail!("signup returned unexpected username");
    }
    if data.signup.token.is_empty() {
        bail!("signup did not return auth token");
    }

    Ok(data.signup)
}

async fn guest_session(
    client: &SmokeClient,
    display_name: &str,
    session_id: &str,
) -> Result<SignupPayload> {
    let data: GuestSessionEnvelope = client
        .post_graphql(
            GUEST_SESSION_MUTATION,
            json!({ "displayName": display_name, "sessionID": session_id }),
            None,
        )
        .await?;

    if data.guest_session.id.is_empty()
        || data.guest_session.username.is_empty()
        || data.guest_session.token.is_empty()
    {
        bail!("guestSession returned an incomplete identity");
    }
    Ok(data.guest_session)
}

async fn create_game(
    client: &SmokeClient,
    user: &SignupPayload,
    game_id: &str,
    session_id: &str,
) -> Result<GamePayload> {
    let payload = json!({
        "ID": game_id,
        "SessionID": session_id,
        "Turn": {
            "Player": user.username,
            "Phase": "MAIN",
            "Number": 1,
            "Priority": user.username,
        },
        "Players": [{
            "UserID": user.id,
            "User": user.username,
            "GameID": game_id,
            "Life": 40,
            "Decklist": CREATOR_DECKLIST,
            "Commander": [],
            "Library": [],
            "Graveyard": [],
            "Exiled": [],
            "Battlefield": [],
            "Hand": [],
            "Revealed": [],
            "Controlled": [],
            "Counters": [],
        }]
    });

    let data: CreateGameEnvelope = client
        .post_graphql(
            CREATE_GAME_MUTATION,
            json!({ "input": payload }),
            Some(&user.token),
        )
        .await?;

    if data.create_game.id != game_id {
        bail!("createGame did not return expected game ID");
    }
    if data.create_game.players.is_empty() {
        bail!("createGame returned no players");
    }
    if data
        .create_game
        .players
        .iter()
        .any(|player| player.id.is_empty() || player.username.is_empty())
    {
        bail!("createGame returned a player missing ID or username");
    }

    Ok(data.create_game)
}

async fn join_game(
    client: &SmokeClient,
    user: &SignupPayload,
    game_id: &str,
    session_id: &str,
) -> Result<GamePayload> {
    let payload = json!({
        "ID": game_id,
        "SessionID": session_id,
        "Decklist": JOINER_DECKLIST,
        "BoardState": {
            "UserID": user.id,
            "User": user.username,
            "GameID": game_id,
            "Life": 40,
            "Commander": [],
            "Library": [],
            "Graveyard": [],
            "Exiled": [],
            "Battlefield": [],
            "Hand": [],
            "Revealed": [],
            "Controlled": [],
            "Counters": [],
        }
    });

    let data: JoinGameEnvelope = client
        .post_graphql(
            JOIN_GAME_MUTATION,
            json!({ "input": payload }),
            Some(&user.token),
        )
        .await?;

    if data.join_game.id != game_id {
        bail!("joinGame did not return expected game ID");
    }
    if data.join_game.players.len() < 2 {
        bail!(
            "joinGame returned {} player(s), expected at least 2",
            data.join_game.players.len()
        );
    }
    if data
        .join_game
        .players
        .iter()
        .any(|player| player.id.is_empty() || player.username.is_empty())
    {
        bail!("joinGame returned a player missing ID or username");
    }

    Ok(data.join_game)
}

async fn game_invite(
    client: &SmokeClient,
    game_id: &str,
    session_id: &str,
) -> Result<GameInvitePayload> {
    let data: GameInviteEnvelope = client
        .post_graphql(
            GAME_INVITE_QUERY,
            json!({ "gameID": game_id, "sessionID": session_id }),
            None,
        )
        .await?;
    let invite = data.game_invite.context("gameInvite returned null")?;
    if invite.id != game_id
        || invite.status != "IN_PROGRESS"
        || invite.player_count != 1
        || invite.capacity < 2
        || invite.player_display_names.is_empty()
    {
        bail!("gameInvite returned an unexpected safe projection");
    }
    Ok(invite)
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    let client = match SmokeClient::new(args.graphql_url.clone(), args.timeout_ms) {
        Ok(client) => client,
        Err(error) => {
            eprintln!("FAIL {}", error);
            std::process::exit(1);
        }
    };

    let result = async {
        let auth_session = format!("smoke-auth-{}", random_suffix());
        let user_a = signup(&client, &format!("smoke_a_{}", random_suffix())).await?;
        let game_id = format!("smoke-game-{}", random_suffix());
        let created = create_game(&client, &user_a, &game_id, &auth_session).await?;
        let user_b = signup(&client, &format!("smoke_b_{}", random_suffix())).await?;
        let joined = join_game(&client, &user_b, &game_id, &auth_session).await?;

        let guest_host_session = format!("smoke-guest-host-{}", random_suffix());
        let guest_host = guest_session(&client, "Smoke Guest Host", &guest_host_session).await?;
        let guest_game_id = format!("smoke-guest-game-{}", random_suffix());
        let guest_created =
            create_game(&client, &guest_host, &guest_game_id, &guest_host_session).await?;
        let guest_join_session = format!("smoke-guest-join-{}", random_suffix());
        let invite = game_invite(&client, &guest_game_id, &guest_join_session).await?;
        let guest_joiner =
            guest_session(&client, "Smoke Guest Invitee", &guest_join_session).await?;
        let guest_joined =
            join_game(&client, &guest_joiner, &guest_game_id, &guest_join_session).await?;
        Ok::<(String, usize, String, usize, usize), anyhow::Error>((
            created.id,
            joined.players.len(),
            guest_created.id,
            guest_joined.players.len(),
            invite.player_display_names.len(),
        ))
    }
    .await;

    match result {
        Ok((game_id, players, guest_game_id, guest_players, invite_names)) => {
            println!(
                "PASS auth_game={} auth_players={} guest_game={} guest_players={} invite_names={} url={}",
                game_id, players, guest_game_id, guest_players, invite_names, args.graphql_url
            );
        }
        Err(error) => {
            eprintln!("FAIL {}", error);
            std::process::exit(1);
        }
    }
}
