package guilds

import (
	"bytes"
	"fmt"
	"github.com/Devs-On-Discord/DoDdy/db"
	bolt "go.etcd.io/bbolt"
)

var (
	guilds                      = []byte("guilds")
	guildId                     = []byte("id")
	guildName                   = []byte("name")
	guildPrefix                 = []byte("prefix")
	guildAnnouncementsChannelID = []byte("announcementsChannelID")
	guildVotesChannelID         = []byte("votesChannelID")
	guildVotes                  = []byte("votes")
)

type Guilds struct {
	db       *bolt.DB
	Prefixes map[string]string
	guilds   map[string]*Guild
}

func (g *Guilds) Init(db *bolt.DB) {
	g.db = db
	g.Prefixes, _ = g.GetPrefixes()
	g.guilds = make(map[string]*Guild)
}

func (g *Guilds) Guild(id string) (*Guild, error) {
	if guild, exists := g.guilds[id]; exists {
		return guild, nil
	} else {
		err := g.db.View(func(tx *bolt.Tx) error {
			guildsBucket := tx.Bucket(guilds)
			if guildsBucket != nil {
				guildBucket := guildsBucket.Bucket([]byte(id))
				if guildBucket != nil {
					guildCursor := guildBucket.Cursor()
					guild = &Guild{db: g.db, id: id, Prefix: ""}
					for k, v := guildCursor.First(); k != nil; k, v = guildCursor.Next() {
						if bytes.Equal(k, guildName) {
							guild.name = string(v)
						} else if bytes.Equal(k, guildPrefix) {
							guild.Prefix = string(v)
						} else if bytes.Equal(k, guildAnnouncementsChannelID) {
							guild.announcementsChannelID = string(v)
						} else if bytes.Equal(k, guildVotesChannelID) {
							guild.votesChannelID = string(v)
						} else if bytes.Equal(k, guildVotes) {
							//TODO: load votes
						}
					}
					g.guilds[guild.id] = guild
					println("guild loaded", guild)
				}
			}
			return nil
		})
		return guild, err
	}
}

func (g *Guild) bucket(tx *bolt.Tx) (*bolt.Bucket, error) {
	guildsBucket := tx.Bucket(guilds)
	if guildsBucket == nil {
		println("guildsBucket==nil")
		return nil, fmt.Errorf(bucketNotCreated)
	}
	guildBucket := guildsBucket.Bucket([]byte(g.id))
	if guildBucket == nil {
		println("id=" + g.id)
		println("guildBucket==nil")
		return nil, fmt.Errorf(notSetup)
	}
	return guildBucket, nil
}

func (g *Guild) SetPrefix(prefix string) (error) {
	err := g.db.Update(func(tx *bolt.Tx) error {
		bucket, err := g.bucket(tx)
		if err != nil {
			return err
		}
		err = bucket.Put(guildPrefix, []byte(prefix))
		if err != nil {
			return fmt.Errorf("prefix couldn't be saved")
		}
		return nil
	})
	if err == nil {
		g.Prefix = prefix
	}
	return err
}

func (g *Guilds) GetPrefixes() (map[string]string, error) {
	prefixes := make(map[string]string)
	err := g.db.View(func(tx *bolt.Tx) error {
		guildsBucket := tx.Bucket([]byte("guilds"))
		if guildsBucket != nil {
			guildsBucket.ForEach(func(k, v []byte) error {
				guildBucket := guildsBucket.Bucket(k)
				if guildBucket != nil {
					prefix := guildBucket.Get([]byte("prefix"))
					if prefix != nil {
						prefixes[string(k)] = string(prefix)
					}
				}
				return nil
			})
		}
		return nil
	})
	return prefixes, err
}

type Guild struct {
	id                     string
	name                   string
	Prefix                 string
	announcementsChannelID string
	votesChannelID         string
	votes                  []GuildVote
	db                     *bolt.DB
}

// GuildVote contains a vote and it's location
type GuildVote struct {
	VoteID    string
	MessageID string
	ChannelID string
}

const (
	notSetup         = "bot isn't set up for this guild"
	bucketNotCreated = "guild's bucket couldn't be created"
)

// GetAnnouncementChannels returns the anouncement channels for every guild
func GetAnnouncementChannels() ([]string, error) {
	channels := make([]string, 0)
	err := db.DB.View(func(tx *bolt.Tx) error {
		guildsBucket := tx.Bucket([]byte("guilds"))
		if guildsBucket == nil {
			return fmt.Errorf(notSetup)
		}
		guildsBucket.ForEach(func(k, v []byte) error {
			guildBucket := guildsBucket.Bucket(k)
			if guildBucket != nil {
				announcementsChannelID := guildBucket.Get([]byte("announcementsChannelID"))
				if announcementsChannelID != nil {
					channels = append(channels, string(announcementsChannelID))
				}
			}
			return nil
		})
		return nil
	})
	return channels, err
}

// SetAnnouncementsChannel sets the anouncement channels for a single guild
func SetAnnouncementsChannel(guildID string, channelID string) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		guildsBucket, err := tx.CreateBucketIfNotExists([]byte("guilds"))
		if err != nil {
			return fmt.Errorf(bucketNotCreated)
		}
		guildBucket := guildsBucket.Bucket([]byte(guildID))
		if guildBucket == nil {
			return fmt.Errorf(notSetup)
		}
		err = guildBucket.Put([]byte("announcementsChannelID"), []byte(channelID))
		if err != nil {
			return fmt.Errorf("announcements channel ID couldn't be saved")
		}
		return nil
	})
}

// GetVotesChannels returns the vote channels for every guild
func GetVotesChannels() ([]string, error) {
	channels := make([]string, 0)
	err := db.DB.View(func(tx *bolt.Tx) error {
		guildsBucket := tx.Bucket([]byte("guilds"))
		if guildsBucket == nil {
			return fmt.Errorf(notSetup)
		}
		guildsBucket.ForEach(func(k, v []byte) error {
			guildBucket := guildsBucket.Bucket(k)
			if guildBucket != nil {
				votesChannelID := guildBucket.Get([]byte("votesChannelID"))
				if votesChannelID != nil {
					channels = append(channels, string(votesChannelID))
				}
			}
			return nil
		})
		return nil
	})
	return channels, err
}

// SetVotesChannel sets the vote channel for a specific guild
func SetVotesChannel(guildID string, channelID string) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		guildsBucket, err := tx.CreateBucketIfNotExists([]byte("guilds"))
		if err != nil {
			return fmt.Errorf(bucketNotCreated)
		}
		guildBucket := guildsBucket.Bucket([]byte(guildID))
		if guildBucket == nil {
			return fmt.Errorf(notSetup)
		}
		err = guildBucket.Put([]byte("votesChannelID"), []byte(channelID))
		if err != nil {
			return fmt.Errorf("votes channel ID couldn't be saved")
		}
		return nil
	})
}

// Create adds a guild to the database
func Create(guildID string, name string) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		guildsBucket, err := tx.CreateBucketIfNotExists([]byte(guilds))
		if err != nil {
			return fmt.Errorf(bucketNotCreated)
		}
		guildBucket, err := guildsBucket.CreateBucket([]byte(guildID))
		if err != nil {
			return fmt.Errorf(bucketNotCreated)
		}
		err = guildBucket.Put([]byte("name"), []byte(name))
		if err != nil {
			return fmt.Errorf("name couldn't be saved")
		}
		err = guildBucket.Put([]byte("prefix"), []byte("!"))
		if err != nil {
			return fmt.Errorf("prefix couldn't be saved")
		}
		return nil
	})
}

// AddVote adds a single vote to a single guild
func AddVote(guildID string, voteID string, messageID string, channelID string) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		guildsBucket, err := tx.CreateBucketIfNotExists([]byte("guilds"))
		if err != nil {
			return fmt.Errorf(bucketNotCreated)
		}
		guildBucket := guildsBucket.Bucket([]byte(guildID))
		if guildBucket == nil {
			return fmt.Errorf(notSetup)
		}
		votesBucket, err := guildBucket.CreateBucketIfNotExists([]byte("votes"))
		if err != nil {
			return fmt.Errorf("vote's bucket couldn't be created")
		}
		voteBucket, err := votesBucket.CreateBucket([]byte(voteID))
		if err != nil {
			return fmt.Errorf("vote bucket couldn't be created")
		}
		err = voteBucket.Put([]byte("voteID"), []byte(voteID))
		if err != nil {
			return fmt.Errorf("voteID couldn't be saved")
		}
		err = voteBucket.Put([]byte("messageID"), []byte(messageID))
		if err != nil {
			return fmt.Errorf("messageID couldn't be saved")
		}
		err = voteBucket.Put([]byte("channelID"), []byte(channelID))
		if err != nil {
			return fmt.Errorf("channelID couldn't be saved")
		}
		return nil
	})
}

// GetVotes returns every single vote from every single guild
func GetVotes() ([]GuildVote, error) {
	votes := make([]GuildVote, 0)
	err := db.DB.View(func(tx *bolt.Tx) error {
		guildsBucket := tx.Bucket([]byte("guilds"))
		if guildsBucket == nil {
			return fmt.Errorf(notSetup)
		}
		err := guildsBucket.ForEach(func(k, v []byte) error {
			guildBucket := guildsBucket.Bucket(k)
			if guildBucket != nil {
				votesBucket := guildBucket.Bucket([]byte("votes"))
				if votesBucket != nil {
					votesBucket.ForEach(func(k, v []byte) error {
						voteBucket := votesBucket.Bucket([]byte(k))
						if voteBucket != nil {
							voteID := voteBucket.Get([]byte("voteID"))
							if voteID == nil {
								return nil
							}
							messageID := voteBucket.Get([]byte("messageID"))
							if messageID == nil {
								return nil
							}
							channelID := voteBucket.Get([]byte("channelID"))
							if channelID == nil {
								return nil
							}
							votes = append(votes, GuildVote{VoteID: string(voteID), MessageID: string(messageID), ChannelID: string(channelID)})
						}
						return nil
					})
				}
			}
			return nil
		})
		return err
	})
	return votes, err
}
