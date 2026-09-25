package model

// These tables were accessed through JDBC in Java, without JPA entities.
type ThematicMaster struct {
	Id                                                                                       int64 `gorm:"primaryKey;autoIncrement"`
	Category, SubCategory, BasketType, BasketName                                            *string
	ShortDescription, LongDescription                                                        *string `gorm:"type:text"`
	Tag, TagDesc, RiskLevel, RiskDescription, Benchmark, Exposure                            *string
	InvestmentDuration, ReviewFrequency, ScripCount, CurrentReturns                          *string
	TotalInvstAmt                                                                            float64
	MinInvstAmnt, OverallWeightage, Methodology, LaunchDate, Rationale, Type, ExpectedReturn *string
	AnalystName, ActionType, Status                                                          *string
	RebalanceAvailable                                                                       int
	IsScheduled, IsSent, IsDraft                                                             int
	ScheduledTime, ExpiryDate                                                                *Timestamp
	CreatedOn                                                                                *Timestamp `gorm:"autoCreateTime"`
	UpdatedOn                                                                                *Timestamp `gorm:"autoUpdateTime"`
	CreatedBy, UpdatedBy                                                                     *string
	ActiveStatus                                                                             int `gorm:"default:1"`
}

func (ThematicMaster) TableName() string { return "tbl_thematic_basket_master" }

type ThematicScrip struct {
	Id                                                                            int64 `gorm:"primaryKey;autoIncrement"`
	BasketId                                                                      int64
	Exchange, Token, Qty, TransType, Price, TradingSymbol, FormattedInsName       *string
	Pdc, Weightage, AdjWeightage, Exposure, OrderType, PriceType, Ret, Source     *string
	TriggerPrice, DisclosedQty, MktProtection, Target, StopLoss, TrailingStopLoss *string
	Version                                                                       int
	CreatedOn                                                                     *Timestamp `gorm:"autoCreateTime"`
	UpdatedOn                                                                     *Timestamp `gorm:"autoUpdateTime"`
	CreatedBy, UpdatedBy                                                          *string
	ActiveStatus                                                                  int `gorm:"default:1"`
}

func (ThematicScrip) TableName() string { return "tbl_thematic_basket_scrips" }

type RebalanceScrip ThematicScrip

func (RebalanceScrip) TableName() string { return "tbl_thematic_basket_rebalance_scrips" }

type ThematicExecution struct {
	Id                                        int64 `gorm:"primaryKey;autoIncrement"`
	UserId                                    string
	BasketId                                  int64
	Lots                                      int
	OrderResponse, OrderRequest, ScripDetails string `gorm:"type:text"`
	IsExecuted, IsViewed                      int
	Source, BasketName, AnalystName, UserName string
	InvestedAmount                            float64
	CreatedBy, UpdatedBy                      string
	CreatedOn                                 *Timestamp `gorm:"autoCreateTime"`
	UpdatedOn                                 *Timestamp `gorm:"autoUpdateTime"`
}

func (ThematicExecution) TableName() string { return "tbl_user_thematic_exec" }

// Columns used by JDBC but absent from the Java JPA declaration.
type ExecutionDetail struct {
	ThematicExecDetails
	ActiveStatus   int `gorm:"default:1"`
	RejectedReason *string
}

func (ExecutionDetail) TableName() string { return "tbl_user_thematic_exec_details" }

type ResearchMaster struct {
	ResearchcallOrderEntity
	AnalystName, Attachement, InvestmentDuration *string
	SortOrder                                    int
	Publish                                      int
}

func (ResearchMaster) TableName() string { return "tbl_researchcall_master" }
