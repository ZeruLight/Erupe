# Erupe

## Client Compatiblity
### Platforms
- PC
- PlayStation 3
- PlayStation Vita
- Wii U (Up to Z2)
### Versions
- ZZ
- Z2
- Z1

## Setup

If you are only looking to install Erupe, please use [a pre-compiled binary](https://github.com/ZeruLight/Erupe/releases/latest).

If you want to modify or compile Erupe yourself, please read on.

## Requirements

- [Go](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)

## Installation

1. Bring up a fresh database by using the [backup file attached with the latest release](https://github.com/ZeruLight/Erupe/releases/latest/download/SCHEMA.sql).
2. Run each script under [patch-schema](./patch-schema) as they introduce newer schema.
3. Edit [config.json](./config.json) such that the database password matches your PostgreSQL setup.
4. Run `go build` or `go run .` to compile Erupe.

## Resources

- [Quest and Scenario Binary Files](https://files.catbox.moe/xf0l7w.7z)
- [PewPewDojo Discord](https://discord.gg/CFnzbhQ)

## Configuration
This portion of the documentation goes over the `config.json` file.

### General Configuration

| Variable               | Description                                                                                                                                           | Default   | Options                         |
|------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------|-----------|---------------------------------|
| Host                   | The IP or host address the server is running from                                                                                                     | 127.0.0.1 |                                 |
| BinPath                | The bin path folder is where you place files needed for various parts of the game such as scenario and quest files                                    | bin       |                                 |
| Language               | This is the language the server will run in. Only English `en` and Japanese `ja` are available, if you wish to contribute to tranlation, get in touch | en        | en/jp                           |
| DisableSoftCrash       |                                                                                                                                                       | false     |                                 |
| HideLoginNotice        | This hides the notices that appear on login from `LoginNotices`                                                                                       | true      |                                 |
| LoginNotices           | This is where you place notices for users, you can have multiple notices                                                                              |           |                                 |
| PatchServerManifest    |                                                                                                                                                       |           |                                 |
| PatchServerFile        |                                                                                                                                                       |           |                                 |
| ScreenshotAPIURL       | This is the URL you want user sreenshots to go to                                                                                                     |           |                                 |
| DeleteOnSaveCorruption | This option deletes a users save from the database if they corrupt it, can be used as punishment for cheaters                                         | false     |                                 |
| ClientMode             | This tells the server what client version it should target                                                                                            | ZZ        | Check compatible versions above |
| DevMode                | This enables DevModeOptions to be configured                                                                                                          | true      |                                 |

### `DevModeOptions` Configuraiton

| Variable             | Description                                                                                 | Default  | Options                          |
|----------------------|---------------------------------------------------------------------------------------------|----------|----------------------------------|
| AutoCreateAccount    | This allows users that don't exist to auto create there account from initial login          | true     |                                  |
| CleanDB              | This cleans the database down                                                               | false    |                                  |
| MaxLauncherHR        | This sets the launcher value to HR7 to allow you to break World HR requirements             | false    |                                  |
| LogInboundMessages   | This will allow inbound messages to be logged to stdout                                     | false    |                                  |
| LogOutboundMessages  | This will allow outbound messages to be logged to stdout                                    | false    |                                  |
| MaxHexdumpLength     | This is the maximum amount of hex bytes that will be dumped to stdout                       | 0        |                                  |
| DivaEvent            | This overrides the Diva event stage in game                                                 | 2        | 0/1/2/3/-1                       |
| FestaEvent           | This overrides the Hunter Festival event stage in game                                      | 2        | 0/1/2/3/-1                       |
| TournamentEvent      | This overrides the Hunter Tournament event stage in game                                    | 2        | 0/1/2/3/-1                       |
| MezFesEvent          | Enables whether the MezFes event & World are active                                         | true     |                                  |
| MezFesAlt            | Switches the multiplayer MezFes event                                                       | false    |                                  |
| DisableTokenCheck    | This disables the random token that is generated at login from being checked, very insecure | false    |                                  |
| QuestDebugTools      | Enable various quest debug logs                                                             | false    |                                  |
| EarthStatusOverride  | Enables Pallone Fest, Tower and Conquest War events                                         | 0        | 2=Conquest, 11=Pallone, 21=Tower |
| EarthIDOverride      | A random event ID                                                                           | 0        |                                  |
| EarthMonsterOverride | Sets the ID of the monster targeted in the Conquest War                                     | 0        |                                  |
| SaveDumps.Enables    | Enables save dumps to a folder that is set at `SaveDumps.OutputDir`                         | true     |                                  |
| SaveDumps.OutputDir  | The folder that save dumps are saved to                                                     | savedata |                                  |

### `GameplayOptions` Configuraiton

| Variable             | Description                                                                 | Default | Options |
|----------------------|-----------------------------------------------------------------------------|---------|---------|
| FeaturedWeapons      | Number of Active Feature weapons to generate daily                          | 0       |         |
| MaximumNP            | Maximum number of NP held by a player                                       | 100000  |         |
| MaximumRP            | Maximum number of RP held by a player                                       | 100000  |         |
| DisableLoginBoost    | Disables the Login Boost system                                             | false   |         |
| DisableBoostTime     | Disables the daily NetCafe Boost Time                                       | false   |         |
| BoostTimeDuration    | The number of minutes NetCafe Boost Time lasts for                          | 120     |         |
| GuildMealDuration    | The number of minutes a Guild Meal can be activated for after cooking       | 60      |         |
| BonusQuestAllowance  | Number of Bonus Point Quests to allow daily                                 | 3       |         |
| DailyQuestAllowance  | Number of Daily Quests to allow daily                                       | 1       |         |
| MezfesSoloTickets    | Number of solo tickets given weekly                                         | 10      |         |
| MezfesGroupTickets   | Number of group tickets given weekly                                        | 4       |         |
| GUrgentRate          | Adjusts the rate of G Urgent quests spawning                                | 10      |         |
| GCPMultiplier        | Adjusts the multiplier of GCP rewarded for quest completion                 | 1.00    |         |
| GRPMultiplier        | Adjusts the multiplier of G Rank Points rewarded for quest completion       | 1.00    |         |
| GSRPMultiplier       | Adjusts the multiplier of G Skill Rank Points rewarded for quest completion | 1.00    |         |
| GZennyMultiplier     | Adjusts the multiplier of G Zenny rewarded for quest completion             | 1.00    |         |
| MaterialMultiplier   | Adjusts the multiplier of Monster Materials rewarded for quest completion   | 1.00    |         |
| ExtraCarves          | Grant n extra chances to carve ALL carcasses                                | 0       |         |
| DisableHunterNavi    | Disables the Hunter Navi                                                    | false   |         |
| EnableHiganjimaEvent | Enables the Higanjima event in the Rasta Bar                                | false   |         |
| EnableNierEvent      | Enables the Nier event in the Rasta Bar                                     | false   |         |
| DisableRoad          | Disables the Hunting Road                                                   | false   |         |

### Discord
There is limited Discord capability in Erupe. The feature allows you to replay messages from your server into a channel.
This may be either be removed or revamped in a future version.

### Commands
There are several chat commands that can be turned on and off. Most of them are really for admins or debugging purposes.

| Name     | command        | Description                                | Options             |
|----------|----------------|--------------------------------------------|---------------------|
| Rights   | !rights VALUE  | Sets the rights integer for your account   |                     |
| Teleport | !tele X,Y      | Teleports user to specific x,y coordinate  |                     |
| Reload   | !reload        | Reloads all users and character objects    |                     |
| KeyQuest | !kqf FLAGS     | Sets the Key Quest Flag for your character |                     |
| Course   | !course OPTION | Enables/Disables a course for your account | HL,EX,Premium,Boost |
| PSN      | !psn  USERNAME | Links the specified PSN to your account    |                     |

### Ravi Sub Commands
| Name     | command                          | Description                   |
|----------|----------------------------------|-------------------------------|
| Raviente | !ravi start                      | Starts Ravi Event             |
| Raviente | !ravi cm / !ravi checkmultiplier | Checks Ravi Damage Multiplier |
| Raviente | !ravi ss                         | Send Sedation Support         |
| Raviente | !ravi sr                         | Send Resurrection Support     |
| Raviente | !ravi rs                         | Request Sedation Support      |


## World `Entries` config

| Config Item | Description      | Options                                                    |
|-------------|------------------|------------------------------------------------------------|
| Type        | Server type.     | 1=Normal, 2=Cities, 3=Newbie, 4=Tavern, 5=Return, 6=MezFes |
| Season      | Server activity. | 0=Green/Breeding, 1=Orange/Warm, 2=Blue/Cold               |

### `Recommend` 
This sets the types of quest that can be ordered from a world.
* 0 = All quests
* 1 = Up to 2 star quests  
* 2 = Up to 4 star quests 
* 4 = All Quests in HR (Enables G Experience Tab) 
* 5 = Only G rank quests 
* 6 = Mini games world there is no place to order quests 
