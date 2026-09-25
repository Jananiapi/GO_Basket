// Generated from Java DTO/entity declarations by tools/generate_models.py; review business logic separately.
package model

// CustomerDTO corresponds to com/sas/dto/CustomerDTO.java.
type CustomerDTO struct {
	UserId              *string         `json:"userId"`
	AccountId           *string         `json:"accountId"`
	Password            *string         `json:"password"`
	NewPassword         *string         `json:"newPassword"`
	OldPassword         *string         `json:"oldPassword"`
	UserSessionID       *string         `json:"userSessionID"`
	UserToken           *string         `json:"userToken"`
	PublicKey4          any             `json:"publicKey4"`
	StringPkey4         *string         `json:"stringPkey4"`
	PublicKey3          any             `json:"publicKey3"`
	StringPkey3         *string         `json:"stringPkey3"`
	Privatekey2         any             `json:"privatekey2"`
	StringPK2           *string         `json:"stringPK2"`
	Tomcatcount         *string         `json:"tomcatcount"`
	Stat                *string         `json:"stat"`
	Tdata               *string         `json:"tdata"`
	SQuestions          *string         `json:"sQuestions"`
	Emsg                *string         `json:"emsg"`
	SCount              *string         `json:"sCount"`
	SIndex              *string         `json:"sIndex"`
	Ques_Ans            *string         `json:"ques_Ans"`
	Answer1             *string         `json:"answer1"`
	Answer2             *string         `json:"answer2"`
	SUserToken          *string         `json:"sUserToken"`
	SPasswordReset      *string         `json:"sPasswordReset"`
	Message             *string         `json:"message"`
	Exch                *string         `json:"exch"`
	Symbol              *string         `json:"symbol"`
	MwName              *string         `json:"mwName"`
	Email               *string         `json:"email"`
	Pan                 *string         `json:"pan"`
	List                *string         `json:"list"`
	PreLogin            *string         `json:"preLogin"`
	UserSettingDto      DefaultLoginDTO `json:"userSettingDto"`
	Is_mob              bool            `json:"is_mob"`
	Dob                 *string         `json:"dob"`
	ScripList           *string         `json:"scripList"`
	ValidAnsResponse    map[string]any  `json:"validAnsResponse"`
	WebSocketID         *string         `json:"webSocketID"`
	AuthorizationStatus int             `json:"authorizationStatus"`
	ExpiryDate          int64           `json:"expiryDate"`
	UserApiKey          *string         `json:"userApiKey"`
	VendorId            int             `json:"vendorId"`
	VendorApiKey        *string         `json:"vendorApiKey"`
	VendorSecret        *string         `json:"vendorSecret"`
	DeviceId            *string         `json:"deviceId"`
	LoginType           *string         `json:"loginType"`
	CallBackUrl         *string         `json:"callBackUrl"`
	LoginMode           *string         `json:"loginMode"`
	Vendor              *string         `json:"vendor"`
	AuthCode            *string         `json:"authCode"`
	Version             *string         `json:"version"`
	IsV2                bool            `json:"isV2"`
	CheckSum            *string         `json:"checkSum"`
	FcmToken            *string         `json:"fcmToken"`
	Imei                *string         `json:"imei"`
}

// DefaultLoginDTO corresponds to com/sas/dto/DefaultLoginDTO.java.
type DefaultLoginDTO struct {
	Stat_DFlogin               *string `json:"stat_DFlogin"`
	Exch                       any     `json:"exch"`
	Prctyp                     any     `json:"prctyp"`
	PCode                      any     `json:"pCode"`
	S_prdt_ali                 any     `json:"s_prdt_ali"`
	STransFlg                  *string `json:"sTransFlg"`
	Default_market_watch_name  *string `json:"default_market_watch_name"`
	Broker_name                *string `json:"broker_name"`
	Branch_id                  *string `json:"branch_id"`
	Market_watch_count         *string `json:"market_watch_count"`
	Email                      *string `json:"email"`
	Weblink                    any     `json:"weblink"`
	Account_id                 *string `json:"account_id"`
	ExchDeatil                 any     `json:"exchDeatil"`
	Lots_weight                *string `json:"lots_weight"`
	Password_special_character *string `json:"password_special_character"`
	AccountName                *string `json:"accountName"`
	UserPrivileges             *string `json:"userPrivileges"`
	YSXorderEntry              *string `json:"ySXorderEntry"`
	CriteriaAttribute_array    any     `json:"criteriaAttribute_array"`
	Emsg_DFlogin               *string `json:"emsg_DFlogin"`
}

// AccessLogModel corresponds to in/codifi/basket/entity/logs/AccessLogModel.java.
type AccessLogModel struct {
	Id           int64      `json:"id"`
	Uri          *string    `json:"uri"`
	Ucc          *string    `json:"ucc"`
	UserId       *string    `json:"userId"`
	ReqId        *string    `json:"reqId"`
	Source       *string    `json:"source"`
	Vendor       *string    `json:"vendor"`
	InTime       *Timestamp `json:"inTime"`
	OutTime      *Timestamp `json:"outTime"`
	LagTime      int64      `json:"lagTime"`
	Module       *string    `json:"module"`
	Method       *string    `json:"method"`
	ReqBody      *string    `json:"reqBody"`
	ResBody      *string    `json:"resBody"`
	DeviceIp     *string    `json:"deviceIp"`
	UserAgent    *string    `json:"userAgent"`
	Domain       *string    `json:"domain"`
	ContentType  *string    `json:"contentType"`
	Session      *string    `json:"session"`
	TableName    *string    `json:"tableName"`
	Elapsed_time *Timestamp `json:"elapsed_time"`
	CreatedOn    *Timestamp `json:"createdOn"`
	UpdatedOn    *Timestamp `json:"updatedOn"`
}

// RestAccessLogModel corresponds to in/codifi/basket/entity/logs/RestAccessLogModel.java.
type RestAccessLogModel struct {
	Id        int64      `json:"id"`
	UserId    *string    `json:"userId"`
	Url       *string    `json:"url"`
	InTime    *Timestamp `json:"inTime"`
	OutTime   *Timestamp `json:"outTime"`
	TotalTime *string    `json:"totalTime"`
	Module    *string    `json:"module"`
	Method    *string    `json:"method"`
	ReqBody   *string    `json:"reqBody"`
	ResBody   *string    `json:"resBody"`
	CreatedOn *Timestamp `json:"createdOn"`
	UpdatedOn *Timestamp `json:"updatedOn"`
}

// BasketNameEntity corresponds to in/codifi/basket/entity/primary/BasketNameEntity.java.
type BasketNameEntity struct {
	UserId       *string             `json:"-" gorm:"column:user_id"`
	BasketId     int64               `json:"basketId" gorm:"column:basket_id;primaryKey;autoIncrement"`
	BasketName   *string             `json:"basketName" gorm:"column:basket_name"`
	Description  *string             `json:"description" gorm:"column:description;type:text"`
	ExpiryDate   *Timestamp          `json:"expiryDate" gorm:"column:expiry_date"`
	ResearchCall int                 `json:"researchCall" gorm:"column:research_call"`
	IsExecuted   *string             `json:"isExecuted" gorm:"column:is_executed;default:0"`
	BasketScrip  []BasketScripEntity `json:"basketScrip" gorm:"foreignKey:BasketId;references:BasketId;constraint:OnDelete:CASCADE"`
	CreatedOn    *Timestamp          `json:"createdOn" gorm:"column:created_on;autoCreateTime"`
	UpdatedOn    *Timestamp          `json:"-" gorm:"column:updated_on;autoUpdateTime"`
	CreatedBy    *string             `json:"-" gorm:"column:created_by"`
	UpdatedBy    *string             `json:"-" gorm:"column:updated_by"`
	ActiveStatus int                 `json:"-" gorm:"column:active_status;default:1"`
}

func (BasketNameEntity) TableName() string { return "tbl_basket_order" }

// BasketScripEntity corresponds to in/codifi/basket/entity/primary/BasketScripEntity.java.
type BasketScripEntity struct {
	Id               int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	BasketId         int64      `json:"basketId" gorm:"column:basket_id"`
	LotSize          *string    `json:"lotSize" gorm:"column:lot_size"`
	SortOrder        *string    `json:"sortOrder" gorm:"column:sort_order"`
	Exchange         *string    `json:"exchange" gorm:"column:exchange"`
	Token            *string    `json:"token" gorm:"column:token"`
	TradingSymbol    *string    `json:"tradingSymbol" gorm:"column:trading_symbol"`
	Qty              *string    `json:"qty" gorm:"column:qty"`
	Price            *string    `json:"price" gorm:"column:price"`
	Expiry           *Timestamp `json:"expiry" gorm:"column:expiry"`
	Product          *string    `json:"product" gorm:"column:product"`
	TransType        *string    `json:"transType" gorm:"column:trans_type"`
	PriceType        *string    `json:"priceType" gorm:"column:price_type"`
	OrderType        *string    `json:"orderType" gorm:"column:order_type"`
	Ret              *string    `json:"ret" gorm:"column:ret"`
	TriggerPrice     *string    `json:"triggerPrice" gorm:"column:trigger_price"`
	DisClosedQty     *string    `json:"disClosedQty" gorm:"column:dis_closed_qty"`
	MktProtection    *string    `json:"mktProtection" gorm:"column:mkt_protection"`
	Target           *string    `json:"target" gorm:"column:target"`
	StopLoss         *string    `json:"stopLoss" gorm:"column:stop_loss"`
	TrailingStopLoss *string    `json:"trailingStopLoss" gorm:"column:trailing_stop_loss"`
	FormattedInsName *string    `json:"formattedInsName" gorm:"column:formatted_ins_name"`
	WeekTag          *string    `json:"weekTag" gorm:"column:week_tag"`
	ValidityDays     *string    `json:"validityDays" gorm:"column:validity_days"`
	ExpiryDate       *string    `json:"expiryDate" gorm:"column:expiry_date"`
	CreatedOn        *Timestamp `json:"-" gorm:"column:created_on;autoCreateTime"`
	UpdatedOn        *Timestamp `json:"-" gorm:"column:updated_on;autoUpdateTime"`
	CreatedBy        *string    `json:"-" gorm:"column:created_by"`
	UpdatedBy        *string    `json:"-" gorm:"column:updated_by"`
	ActiveStatus     int        `json:"-" gorm:"column:active_status;default:1"`
}

func (BasketScripEntity) TableName() string { return "tbl_basket_order_scrip" }

// CommonEntity corresponds to in/codifi/basket/entity/primary/CommonEntity.java.
type CommonEntity struct {
	Id           int64      `json:"-" gorm:"column:id;primaryKey;autoIncrement"`
	CreatedOn    *Timestamp `json:"createdOn" gorm:"column:created_on;autoCreateTime"`
	UpdatedOn    *Timestamp `json:"-" gorm:"column:updated_on;autoUpdateTime"`
	CreatedBy    *string    `json:"-" gorm:"column:created_by"`
	UpdatedBy    *string    `json:"-" gorm:"column:updated_by"`
	ActiveStatus int        `json:"-" gorm:"column:active_status;default:1"`
}

// DeviceMappingEntity corresponds to in/codifi/basket/entity/primary/DeviceMappingEntity.java.
type DeviceMappingEntity struct {
	CommonEntity
	UserName   *string `json:"userName" gorm:"column:user_name"`
	UserId     *string `json:"userId" gorm:"column:user_id"`
	DeviceId   *string `json:"deviceId" gorm:"column:device_id"`
	DeviceType *string `json:"deviceType" gorm:"column:device_type"`
}

func (DeviceMappingEntity) TableName() string { return "tbl_device_mapping" }

// OrderStatusFeedEntity corresponds to in/codifi/basket/entity/primary/OrderStatusFeedEntity.java.
type OrderStatusFeedEntity struct {
	Id             int64      `json:"-" gorm:"column:id;primaryKey;autoIncrement"`
	UserId         *string    `json:"userId" gorm:"column:user_id"`
	OrderNo        *string    `json:"orderNo" gorm:"column:order_no"`
	Token          *string    `json:"token" gorm:"column:token"`
	Exch           *string    `json:"exch" gorm:"column:exch"`
	TradingSymbol  *string    `json:"tradingSymbol" gorm:"column:trading_symbol"`
	TransType      *string    `json:"transType" gorm:"column:trans_type"`
	Price          *string    `json:"price" gorm:"column:price"`
	TradedPrice    *string    `json:"tradedPrice" gorm:"column:traded_price"`
	TriggerPrice   *string    `json:"triggerPrice" gorm:"column:trigger_price"`
	Qty            *string    `json:"qty" gorm:"column:qty"`
	PendingQty     *string    `json:"pendingQty" gorm:"column:pending_qty"`
	TradedQty      *string    `json:"tradedQty" gorm:"column:traded_qty"`
	OrderType      *string    `json:"orderType" gorm:"column:order_type"`
	InstrumentName *string    `json:"instrumentName" gorm:"column:instrument_name"`
	OrderStatus    *string    `json:"orderStatus" gorm:"column:order_status"`
	Reason         *string    `json:"reason" gorm:"column:reason"`
	UserRemarks    *string    `json:"userRemarks" gorm:"column:user_remarks"`
	OrderEntryTime *string    `json:"orderEntryTime" gorm:"column:order_entry_time"`
	Response       *string    `json:"response" gorm:"column:response;type:text"`
	Vendor         *string    `json:"vendor" gorm:"column:vendor"`
	CreatedOn      *Timestamp `json:"-" gorm:"column:created_on;autoCreateTime"`
	UpdatedOn      *Timestamp `json:"-" gorm:"column:updated_on;autoUpdateTime"`
	CreatedBy      *string    `json:"-" gorm:"column:created_by"`
	UpdatedBy      *string    `json:"-" gorm:"column:updated_by"`
	ActiveStatus   int        `json:"-" gorm:"column:active_status;default:1"`
}

func (OrderStatusFeedEntity) TableName() string { return "tbl_order_status_feed" }

// ReasearchCallUsers corresponds to in/codifi/basket/entity/primary/ReasearchCallUsers.java.
type ReasearchCallUsers struct {
	Id               int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ResearchCallId   int        `json:"researchCallId" gorm:"column:researchcall_id"`
	ThematicBasketId int        `json:"thematicBasketId" gorm:"column:thematic_basket_id"`
	StrategyId       int        `json:"strategyId" gorm:"column:strategy_id"`
	UserId           *string    `json:"userId" gorm:"column:user_id"`
	ActiveStatus     int        `json:"activeStatus" gorm:"column:active_status"`
	CreatedBy        *string    `json:"createdBy" gorm:"column:created_by"`
	UpdatedBy        *string    `json:"updatedBy" gorm:"column:updated_by"`
	CreatedOn        *Timestamp `json:"-" gorm:"column:created_on;autoUpdateTime"`
	UpdatedOn        *Timestamp `json:"-" gorm:"column:updated_on;autoUpdateTime"`
}

func (ReasearchCallUsers) TableName() string { return "tbl_researchcall_usermapping" }

// ResearchCallStatusEntity corresponds to in/codifi/basket/entity/primary/ResearchCallStatusEntity.java.
type ResearchCallStatusEntity struct {
	CommonEntity
	StatusCode int     `json:"statusCode" gorm:"column:statuscode"`
	Status     *string `json:"status" gorm:"column:status"`
}

func (ResearchCallStatusEntity) TableName() string { return "tbl_researchcall_status" }

// ResearchcallOrderEntity corresponds to in/codifi/basket/entity/primary/ResearchcallOrderEntity.java.
type ResearchcallOrderEntity struct {
	Id                    int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	BasketName            *string    `json:"basketName" gorm:"column:basket_name"`
	UserId                *string    `json:"-" gorm:"column:user_id"`
	ExpiryDate            *Timestamp `json:"expiryDate" gorm:"column:expiry_date"`
	Category              *string    `json:"category" gorm:"column:category"`
	SubCategory           *string    `json:"subCategory" gorm:"column:subcategory"`
	Tags                  *string    `json:"tags" gorm:"column:tags"`
	ShortDescription      *string    `json:"shortDescription" gorm:"column:shortdescription;type:text"`
	LongDescription       *string    `json:"longDescription" gorm:"column:longdescription;type:text"`
	IsExecuted            *string    `json:"isExecuted" gorm:"column:is_executed;default:0"`
	IsVendorBasket        int        `json:"isVendorBasket" gorm:"column:is_vendor_basket"`
	ResearchCall          int        `json:"researchCall" gorm:"column:researchcall"`
	VendorCode            *string    `json:"vendorCode" gorm:"column:vendor_code"`
	SendPushNotification  int        `json:"sendPushNotification" gorm:"column:send_pushnotification"`
	PushNotificationTitle *string    `json:"pushNotificationTitle" gorm:"column:pushnotification_title"`
	Source                *string    `json:"source" gorm:"column:source"`
	Remarks               *string    `json:"remarks" gorm:"column:remarks"`
	Channels              *string    `json:"channels" gorm:"column:channels"`
	Status                *string    `json:"status" gorm:"column:status"`
	SpeclizationTag       *string    `json:"speclizationTag" gorm:"column:speclization_tag"`
	ActiveStatus          int        `json:"-" gorm:"column:active_status;default:1"`
	UpdatedBy             *string    `json:"-" gorm:"column:updated_by"`
	UpdatedOn             *Timestamp `json:"-" gorm:"column:updated_on;autoUpdateTime"`
	CreatedBy             *string    `json:"-" gorm:"column:created_by"`
	CreatedOn             *Timestamp `json:"createdOn" gorm:"column:created_on;autoCreateTime"`
}

func (ResearchcallOrderEntity) TableName() string { return "tbl_researchcall_master" }

// ResearchcallScripEntity corresponds to in/codifi/basket/entity/primary/ResearchcallScripEntity.java.
type ResearchcallScripEntity struct {
	Id                 int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ResearchcallId     int        `json:"researchcallId" gorm:"column:researchcall_id"`
	Token              *string    `json:"token" gorm:"column:token"`
	Exchange           *string    `json:"exchange" gorm:"column:exchange"`
	Validity           *string    `json:"validity" gorm:"column:validity"`
	Qty                *string    `json:"qty" gorm:"column:qty"`
	Product            *string    `json:"product" gorm:"column:product"`
	TransType          *string    `json:"transType" gorm:"column:trans_type"`
	PriceType          *string    `json:"priceType" gorm:"column:price_type"`
	OrderType          *string    `json:"orderType" gorm:"column:order_type"`
	PriceLowerBound    *string    `json:"priceLowerBound" gorm:"column:price_lower_bound"`
	PriceUpperBound    *string    `json:"priceUpperBound" gorm:"column:price_upper_bound"`
	StopLossUpperBound *string    `json:"stopLossUpperBound" gorm:"column:stoploss_upper_bound"`
	StopLossLowerBound *string    `json:"stopLossLowerBound" gorm:"column:stoploss_lower_bound"`
	TargetUpperBound   *string    `json:"targetUpperBound" gorm:"column:target_upper_bound"`
	TargetLowerBound   *string    `json:"targetLowerBound" gorm:"column:target_lower_bound"`
	Retention          *string    `json:"retention" gorm:"column:retention"`
	TrailingStopLoss   *string    `json:"trailingStopLoss" gorm:"column:trailing_stop_loss"`
	TriggerPrice       *string    `json:"triggerPrice" gorm:"column:trigger_price"`
	DisClosedQty       *string    `json:"disClosedQty" gorm:"column:dis_closed_qty"`
	Expiry             *string    `json:"expiry" gorm:"column:expiry"`
	ExpiryDate         *string    `json:"expiryDate" gorm:"column:expiry_date"`
	ValidityDays       *string    `json:"validityDays" gorm:"column:validity_days"`
	FormattedInsName   *string    `json:"formattedInsName" gorm:"column:formatted_ins_name"`
	TradingSymbol      *string    `json:"tradingSymbol" gorm:"column:trading_symbol"`
	WeekTag            *string    `json:"weekTag" gorm:"column:week_tag"`
	Remarks            *string    `json:"remarks" gorm:"column:remarks"`
	LotSize            *string    `json:"lotSize" gorm:"column:lot_size"`
	MktProtection      *string    `json:"mktProtection" gorm:"column:mkt_protection"`
	UserSegment        *string    `json:"userSegment" gorm:"column:user_segment"`
	ActiveStatus       int        `json:"-" gorm:"column:active_status"`
	CreatedOn          *Timestamp `json:"-" gorm:"column:created_on;autoCreateTime"`
	UpdatedOn          *Timestamp `json:"-" gorm:"column:updated_on;autoUpdateTime"`
	CreatedBy          *string    `json:"-" gorm:"column:created_by"`
	UpdatedBy          *string    `json:"-" gorm:"column:updated_by"`
}

func (ResearchcallScripEntity) TableName() string { return "tbl_research_scrip" }

// SectorReportsEntity corresponds to in/codifi/basket/entity/primary/SectorReportsEntity.java.
type SectorReportsEntity struct {
	Id           *int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ReferanceId  *int64     `json:"referanceId" gorm:"column:referance_id"`
	Name         *string    `json:"name" gorm:"column:name"`
	Description  *string    `json:"description" gorm:"column:description;type:text"`
	Attachment   *string    `json:"attachment" gorm:"column:attachment;type:text"`
	Url          *string    `json:"url" gorm:"column:url"`
	Type         *string    `json:"type" gorm:"column:type"`
	UpdatedBy    *string    `json:"updatedBy" gorm:"column:updated_by"`
	UpdatedOn    *Timestamp `json:"-" gorm:"column:updated_on;autoUpdateTime"`
	CreatedBy    *string    `json:"createdBy" gorm:"column:created_by"`
	CreatedOn    *Timestamp `json:"-" gorm:"column:created_on;autoCreateTime"`
	ActiveStatus int        `json:"activeStatus" gorm:"column:active_status"`
}

func (SectorReportsEntity) TableName() string { return "tbl_sector_reports" }

// ThematicExeMasterEntity corresponds to in/codifi/basket/entity/primary/ThematicExeMasterEntity.java.
type ThematicExeMasterEntity struct {
	Id           int64      `json:"-" gorm:"column:id;primaryKey;autoIncrement"`
	UserId       *string    `json:"-" gorm:"column:user_id"`
	BasketId     int64      `json:"basketId" gorm:"column:basket_id"`
	ResearchType int64      `json:"researchType" gorm:"column:research_type"`
	Lots         *int       `json:"lots" gorm:"column:lots"`
	BasketName   *string    `json:"basketName" gorm:"column:basket_name"`
	AnalystName  *string    `json:"analystName" gorm:"column:analyst_name"`
	BasketAction *string    `json:"basketAction" gorm:"column:basket_action"`
	Version      *string    `json:"version" gorm:"column:version"`
	Source       *string    `json:"source" gorm:"column:source"`
	CreatedOn    *Timestamp `json:"createdOn" gorm:"column:created_on;autoCreateTime"`
	UpdatedOn    *Timestamp `json:"-" gorm:"column:updated_on;autoUpdateTime"`
	CreatedBy    *string    `json:"-" gorm:"column:created_by"`
	UpdatedBy    *string    `json:"-" gorm:"column:updated_by"`
	ActiveStatus int        `json:"-" gorm:"column:active_status;default:1"`
}

func (ThematicExeMasterEntity) TableName() string { return "tbl_user_thematic_master" }

// ThematicExecDetails corresponds to in/codifi/basket/entity/primary/ThematicExecDetails.java.
type ThematicExecDetails struct {
	Id            *int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ExecutionId   *string    `json:"executionId" gorm:"column:execution_id"`
	UserId        *string    `json:"userId" gorm:"column:user_id"`
	ResearchId    *int       `json:"researchId" gorm:"column:research_id"`
	BasketName    *string    `json:"basketName" gorm:"column:basket_name"`
	ResearchType  *int       `json:"researchType" gorm:"column:research_type"`
	LotSize       *int       `json:"lotSize" gorm:"column:lot_size"`
	UserPrice     *float32   `json:"userPrice" gorm:"column:user_price"`
	ExecutedPrice *string    `json:"executedPrice" gorm:"column:executed_price"`
	Exch          *string    `json:"exch" gorm:"column:exch"`
	Token         *string    `json:"token" gorm:"column:token"`
	TradingSymbol *string    `json:"tradingSymbol" gorm:"column:trading_symbol"`
	OrderNo       *string    `json:"orderNo" gorm:"column:order_no"`
	OrderResponse *string    `json:"orderResponse" gorm:"column:order_response;type:text"`
	ExecutedOn    *Timestamp `json:"executedOn" gorm:"column:executed_on"`
	RecommQty     *int       `json:"recommQty" gorm:"column:recomm_qty"`
	OriginalQty   *int       `json:"originalQty" gorm:"column:original_qty"`
	ExecutedQty   *int       `json:"executedQty" gorm:"column:executed_qty"`
	OrderStatus   *string    `json:"orderStatus" gorm:"column:order_status"`
	IsSuccessful  *bool      `json:"isSuccessful" gorm:"column:is_successful"`
	CreatedBy     *string    `json:"createdBy" gorm:"column:created_by"`
	CreatedOn     *Timestamp `json:"createdOn" gorm:"column:created_on"`
	UpdatedBy     *string    `json:"updatedBy" gorm:"column:updated_by"`
	UpdatedOn     *Timestamp `json:"updatedOn" gorm:"column:updated_on"`
	RetryCount    *int       `json:"retryCount" gorm:"column:retry_count"`
	Version       *int       `json:"version" gorm:"column:version"`
	TransType     *string    `json:"transType" gorm:"column:trans_type"`
	BasketAction  *string    `json:"basketAction" gorm:"column:basket_action"`
}

func (ThematicExecDetails) TableName() string { return "tbl_user_thematic_exec_details" }

// UserNotification corresponds to in/codifi/basket/entity/primary/UserNotification.java.
type UserNotification struct {
	Id                  int        `json:"-" gorm:"column:id;primaryKey;autoIncrement"`
	Message             *string    `json:"message" gorm:"column:message;type:text"`
	Url                 *string    `json:"url" gorm:"column:url"`
	Title               *string    `json:"title" gorm:"column:title"`
	BasketId            *string    `json:"basketId" gorm:"column:basket_id"`
	MessageType         *string    `json:"messageType" gorm:"column:message_type;type:text"`
	UserId              *string    `json:"userId" gorm:"column:user_id"`
	UserType            *string    `json:"userType" gorm:"column:user_type"`
	Icon                *string    `json:"icon" gorm:"column:icon"`
	Validity            *Timestamp `json:"validity" gorm:"column:validity"`
	OrderRecommendation *string    `json:"orderRecommendation" gorm:"column:order_recommendation;type:text"`
	CreatedOn           *Timestamp `json:"-" gorm:"column:created_on;autoCreateTime"`
	CreatedBy           *string    `json:"createdBy" gorm:"column:created_by"`
	UpdatedOn           *Timestamp `json:"-" gorm:"column:updated_on;autoUpdateTime"`
	UpdatedBy           *string    `json:"updatedBy" gorm:"column:updated_by"`
	ActiveStatus        *bool      `json:"activeStatus" gorm:"column:active_status"`
}

func (UserNotification) TableName() string { return "tbl_user_notification" }

// VendorAppEntity corresponds to in/codifi/basket/entity/primary/VendorAppEntity.java.
type VendorAppEntity struct {
	Id                   int     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	AppName              *string `json:"appName" gorm:"column:app_name"`
	ApiKey               *string `json:"apiKey" gorm:"column:api_key"`
	ApiSecret            *string `json:"apiSecret" gorm:"column:api_secret"`
	TppAuthorization     int     `json:"tppAuthorization" gorm:"column:tpp_authorization"`
	Client_id            *string `json:"client_id" gorm:"column:client_id"`
	Contact_name         *string `json:"contact_name" gorm:"column:contact_name"`
	Mobile_no            *string `json:"mobile_no" gorm:"column:mobile_no"`
	Email                *string `json:"email" gorm:"column:email"`
	Type                 *string `json:"type" gorm:"column:type"`
	Icon_url             *string `json:"icon_url" gorm:"column:icon_url"`
	Redirect_url         *string `json:"redirect_url" gorm:"column:redirect_url"`
	Postback_url         *string `json:"postback_url" gorm:"column:postback_url"`
	Description          *string `json:"description" gorm:"column:description;type:text"`
	Authorization_status int     `json:"authorization_status" gorm:"column:authorization_status"`
	Is_accepted          int     `json:"is_accepted" gorm:"column:is_accepted"`
	Rejected_reson       *string `json:"rejected_reson" gorm:"column:rejected_reson"`
}

func (VendorAppEntity) TableName() string { return "tbl_vendor_app" }

// AdminBasketOrderReq corresponds to in/codifi/basket/model/request/AdminBasketOrderReq.java.
type AdminBasketOrderReq struct {
	ApiKey           *string             `json:"apiKey"`
	BasketName       *string             `json:"basketName"`
	UserId           []string            `json:"userId"`
	BasketId         int                 `json:"basketId"`
	Description      *string             `json:"description"`
	ExpiryDate       *Timestamp          `json:"expiryDate"`
	PushNotification int                 `json:"pushNotification"`
	Message          *string             `json:"message"`
	Title            *string             `json:"title"`
	Scrips           []ScripRequestModel `json:"scrips"`
	ScripsId         []int64             `json:"scripsId"`
}

// BasketMarginRequest corresponds to in/codifi/basket/model/request/BasketMarginRequest.java.
type BasketMarginRequest struct {
	Exchange  *string `json:"exchange"`
	Symbol    *string `json:"symbol"`
	Qty       *string `json:"qty"`
	Token     *string `json:"token"`
	Price     *string `json:"price"`
	TransType *string `json:"transType"`
}

// BasketOrderListUpdateReq corresponds to in/codifi/basket/model/request/BasketOrderListUpdateReq.java.
type BasketOrderListUpdateReq struct {
	BasketName *string             `json:"basketName"`
	BasketId   int                 `json:"basketId"`
	Scrips     []ScripRequestModel `json:"scrips"`
	ScripsId   []int64             `json:"scripsId"`
}

// BasketOrderReq corresponds to in/codifi/basket/model/request/BasketOrderReq.java.
type BasketOrderReq struct {
	BasketName *string           `json:"basketName"`
	BasketId   int               `json:"basketId"`
	Scrips     ScripRequestModel `json:"scrips"`
	ScripsId   []int64           `json:"scripsId"`
}

// ExecuteBasketOrderReq corresponds to in/codifi/basket/model/request/ExecuteBasketOrderReq.java.
type ExecuteBasketOrderReq struct {
	BasketName       *string             `json:"basketName"`
	BasketId         int                 `json:"basketId"`
	Lots             int                 `json:"lots"`
	LotSize          int                 `json:"lotSize"`
	Source           *string             `json:"source"`
	InvestmentAmount float64             `json:"investmentAmount"`
	IsModified       int                 `json:"isModified"`
	IsViewed         int                 `json:"isViewed"`
	BasketAction     *string             `json:"basketAction"`
	Scrips           []ScripRequestModel `json:"scrips"`
}

// OrderBookReqModel corresponds to in/codifi/basket/model/request/OrderBookReqModel.java.
type OrderBookReqModel struct {
	Uid *string `json:"uid"`
	Prd *string `json:"prd"`
}

// ResearchCallModelRequest corresponds to in/codifi/basket/model/request/ResearchCallModelRequest.java.
type ResearchCallModelRequest struct {
	ClientId              *string `json:"clientId"`
	ResearchCall          *string `json:"researchCall"`
	BasketId              int64   `json:"basketId"`
	ExpiryDate            *string `json:"expiryDate"`
	Status                *string `json:"status"`
	Source                *string `json:"source"`
	SendPushNotification  int     `json:"sendPushNotification"`
	PushNotificationTitle *string `json:"pushNotificationTitle"`
}

// ResearchCallRequest corresponds to in/codifi/basket/model/request/ResearchCallRequest.java.
type ResearchCallRequest struct {
	Id             int64    `json:"id"`
	ClientId       *string  `json:"clientId"`
	BasketId       *string  `json:"basketId"`
	ActiveStatus   *string  `json:"activeStatus"`
	ResearchCall   *string  `json:"researchCall"`
	ResearchCallId *int64   `json:"researchCallId"`
	Status         *string  `json:"status"`
	Category       *string  `json:"category"`
	SubCategory    *string  `json:"subCategory"`
	Tags           []string `json:"tags"`
	FromDate       *string  `json:"fromDate"`
	ToDate         *string  `json:"toDate"`
	Remarks        *string  `json:"remarks"`
	AnalystName    *string  `json:"analystName"`
	BasketType     *string  `json:"basketType"`
}

// RetrieveBasketModel corresponds to in/codifi/basket/model/request/RetrieveBasketModel.java.
type RetrieveBasketModel struct {
	BasketId   int64      `json:"basketId"`
	BasketName *string    `json:"basketName"`
	IsExecuted *string    `json:"isExecuted" gorm:"default:0"`
	CreatedOn  *Timestamp `json:"createdOn"`
	ScripCount int64      `json:"scripCount"`
}

// ScripRequest corresponds to in/codifi/basket/model/request/ScripRequest.java.
type ScripRequest struct {
	Exchange  *string `json:"exchange"`
	Token     *string `json:"token"`
	Qty       *string `json:"qty"`
	TransType *string `json:"transType"`
	Price     *string `json:"price"`
}

// ScripRequestModel corresponds to in/codifi/basket/model/request/ScripRequestModel.java.
type ScripRequestModel struct {
	Id               int64      `json:"id"`
	SortOrder        *string    `json:"sortOrder"`
	Exchange         *string    `json:"exchange"`
	Token            *string    `json:"token"`
	TradingSymbol    *string    `json:"tradingSymbol"`
	Qty              *string    `json:"qty"`
	Price            *string    `json:"price"`
	Product          *string    `json:"product"`
	TransType        *string    `json:"transType"`
	PriceType        *string    `json:"priceType"`
	OrderType        *string    `json:"orderType"`
	Ret              *string    `json:"ret"`
	TriggerPrice     *string    `json:"triggerPrice"`
	DisClosedQty     *string    `json:"disClosedQty"`
	MktProtection    *string    `json:"mktProtection"`
	Target           *string    `json:"target"`
	StopLoss         *string    `json:"stopLoss"`
	TrailingStopLoss *string    `json:"trailingStopLoss"`
	CreatedBy        *string    `json:"createdBy"`
	Source           *string    `json:"source"`
	LotSize          *string    `json:"lotSize"`
	FormattedInsName *string    `json:"formattedInsName"`
	WeekTag          *string    `json:"weekTag"`
	Expiry           *Timestamp `json:"expiry"`
	ExpiryDate       *string    `json:"expiryDate"`
	ValidityDays     *string    `json:"validityDays"`
	Ltp              float64    `json:"ltp"`
	UserSegment      *string    `json:"userSegment"`
	Weightage        *string    `json:"weightage"`
	Version          *int       `json:"version"`
}

// SendNoficationReqModel corresponds to in/codifi/basket/model/request/SendNoficationReqModel.java.
type SendNoficationReqModel struct {
	Message             *string        `json:"message"`
	Url                 *string        `json:"url"`
	Title               *string        `json:"title"`
	MessageType         *string        `json:"messageType"`
	UserId              []string       `json:"userId"`
	UserType            *string        `json:"userType"`
	OrderRecommendation map[string]any `json:"orderRecommendation"`
	Icon                *string        `json:"icon"`
	Validity            *Timestamp     `json:"validity"`
	BasketId            *string        `json:"basketId"`
}

// SpanMarginReq corresponds to in/codifi/basket/model/request/SpanMarginReq.java.
type SpanMarginReq struct {
	Token     *string `json:"token"`
	Exchange  *string `json:"exchange"`
	Price     *string `json:"price"`
	Qty       *string `json:"qty"`
	TransType *string `json:"transType"`
}

// ThematicBasketMaster corresponds to in/codifi/basket/model/request/ThematicBasketMaster.java.
type ThematicBasketMaster struct {
	Id                 int        `json:"id"`
	Category           *string    `json:"category"`
	SubCategory        *string    `json:"subCategory"`
	BasketType         *string    `json:"basketType"`
	BasketName         *string    `json:"basketName"`
	ShortDescription   *string    `json:"shortDescription"`
	LongDescription    *string    `json:"longDescription"`
	ScripCount         *string    `json:"scripCount"`
	TotalInvstAmt      *string    `json:"totalInvstAmt"`
	CurrentReturns     *string    `json:"currentReturns"`
	Tag                *string    `json:"tag"`
	RiskLevel          *string    `json:"riskLevel"`
	Benchmark          *string    `json:"benchmark"`
	Exposure           *string    `json:"exposure"`
	InvestmentDuration *string    `json:"investmentDuration"`
	ReviewFrequency    *string    `json:"reviewFrequency"`
	RebalanceAvailable bool       `json:"rebalanceAvailable"`
	CreatedBy          *string    `json:"createdBy"`
	UpdatedBy          *string    `json:"updatedBy"`
	ActiveStatus       bool       `json:"activeStatus"`
	AnalystName        *string    `json:"analystName"`
	Scheduled          bool       `json:"scheduled"`
	ScheduledTime      *Timestamp `json:"scheduledTime"`
	Sent               bool       `json:"sent"`
	Draft              bool       `json:"draft"`
	User               *string    `json:"user"`
}

// ThematicBasketRequest corresponds to in/codifi/basket/model/request/ThematicBasketRequest.java.
type ThematicBasketRequest struct {
	Category           *string        `json:"category"`
	SubCategory        *string        `json:"subCategory"`
	BasketName         *string        `json:"basketName"`
	ShortDescription   *string        `json:"shortDescription"`
	LongDescription    *string        `json:"longDescription"`
	Tag                *string        `json:"tag"`
	RiskLevel          *string        `json:"riskLevel"`
	Benchmark          *string        `json:"benchmark"`
	Exposure           *string        `json:"exposure"`
	InvestmentDuration *string        `json:"investmentDuration"`
	ReviewFrequency    *string        `json:"reviewFrequency"`
	ScripCount         *string        `json:"scripCount"`
	Scrips             []ScripRequest `json:"scrips"`
}

// ThematicBasketScrip corresponds to in/codifi/basket/model/request/ThematicBasketScrip.java.
type ThematicBasketScrip struct {
	Id               int64   `json:"id"`
	BasketId         int     `json:"basketId"`
	Exchange         *string `json:"exchange"`
	Token            *string `json:"token"`
	Qty              *int    `json:"qty"`
	Weightage        *string `json:"weightage"`
	AdjWeightage     *string `json:"adjWeightage"`
	TransType        *string `json:"transType"`
	Price            *string `json:"price"`
	Pdc              *string `json:"pdc"`
	TradingSymbol    *string `json:"tradingSymbol"`
	FormattedInsName *string `json:"formattedInsName"`
	OrderType        *string `json:"orderType"`
	PriceType        *string `json:"priceType"`
	CreatedBy        *string `json:"createdBy"`
	UpdatedBy        *string `json:"updatedBy"`
}

// AdminRecommendation corresponds to in/codifi/basket/model/response/AdminRecommendation.java.
type AdminRecommendation struct {
	Token         *string `json:"token"`
	Exchange      *string `json:"exchange"`
	TradingSymbol *string `json:"tradingSymbol"`
	Qty           int     `json:"qty"`
	OrgRecoQty    int     `json:"orgRecoQty"`
}

// BasketHoldingsResponse corresponds to in/codifi/basket/model/response/BasketHoldingsResponse.java.
type BasketHoldingsResponse struct {
	BasketId          int                 `json:"basketId"`
	BasketName        *string             `json:"basketName"`
	UserId            *string             `json:"userId"`
	InvestedAmount    float64             `json:"investedAmount"`
	CurrentVersion    int                 `json:"currentVersion"`
	RebalancedVersion int                 `json:"rebalancedVersion"`
	IsRebalanced      int                 `json:"isRebalanced"`
	LotSize           int                 `json:"lotSize"`
	ExecutedDate      *string             `json:"executedDate"`
	ScripList         []ScripHoldingModel `json:"scripList"`
}

// BasketMasterModel corresponds to in/codifi/basket/model/response/BasketMasterModel.java.
type BasketMasterModel struct {
	BasketId       int        `json:"basketId"`
	BasketName     *string    `json:"basketName"`
	InvestedAmount float64    `json:"investedAmount"`
	CreatedDate    *string    `json:"createdDate"`
	LotSize        *string    `json:"lotSize"`
	Status         int        `json:"status"`
	CreatedBy      *string    `json:"createdBy"`
	ModifiedBy     *string    `json:"modifiedBy"`
	ModifiedDate   *string    `json:"modifiedDate"`
	Id             int        `json:"id"`
	TotalInvstAmt  *string    `json:"totalInvstAmt"`
	CreatedOn      *Timestamp `json:"createdOn"`
}

// BasketScripModel corresponds to in/codifi/basket/model/response/BasketScripModel.java.
type BasketScripModel struct {
	Token         *string `json:"token"`
	Exchange      *string `json:"exchange"`
	RecommQty     int     `json:"recommQty"`
	Qty           int     `json:"qty"`
	TradingSymbol *string `json:"tradingSymbol"`
}

// GenericOrderBookResp corresponds to in/codifi/basket/model/response/GenericOrderBookResp.java.
type GenericOrderBookResp struct {
	OrderNo          *string `json:"orderNo"`
	UserId           *string `json:"userId"`
	ActId            *string `json:"actId"`
	Exchange         *string `json:"exchange"`
	CompanyName      *string `json:"companyName"`
	TradingSymbol    *string `json:"tradingSymbol"`
	Qty              *string `json:"qty"`
	TransType        *string `json:"transType"`
	Ret              *string `json:"ret"`
	Token            *string `json:"token"`
	Multiplier       *string `json:"multiplier"`
	LotSize          *string `json:"lotSize"`
	TickSize         *string `json:"tickSize"`
	Price            *string `json:"price"`
	RPrice           *string `json:"rPrice"`
	AvgTradePrice    *string `json:"avgTradePrice"`
	AvgPrc           *string `json:"avgPrc"`
	Prc              *string `json:"prc"`
	DisclosedQty     *string `json:"disclosedQty"`
	Product          *string `json:"product"`
	PriceType        *string `json:"priceType"`
	OrderType        *string `json:"orderType"`
	OrderStatus      *string `json:"orderStatus"`
	FillShares       *string `json:"fillShares"`
	ExchUpdateTime   *string `json:"exchUpdateTime"`
	ExchOrderId      *string `json:"exchOrderId"`
	RQty             *string `json:"rQty"`
	FormattedInsName *string `json:"formattedInsName"`
	Ltp              *string `json:"ltp"`
	RejectedReason   *string `json:"rejectedReason"`
	TriggerPrice     *string `json:"triggerPrice"`
	MktProtection    *string `json:"mktProtection"`
	Target           *string `json:"target"`
	StopLoss         *string `json:"stopLoss"`
	TrailingPrice    *string `json:"trailingPrice"`
	OrderTime        *string `json:"orderTime"`
	InitiatedBy      *string `json:"initiatedBy"`
	Remark           *string `json:"remark"`
}

// GenericOrderResp corresponds to in/codifi/basket/model/response/GenericOrderResp.java.
type GenericOrderResp struct {
	RequestTime *string `json:"requestTime"`
	OrderNo     *string `json:"orderNo"`
}

// GenericResponse corresponds to in/codifi/basket/model/response/GenericResponse.java.
type GenericResponse struct {
	Status  *string `json:"status"`
	Message *string `json:"message"`
	Result  any     `json:"result"`
}

// ResearchCallResponse corresponds to in/codifi/basket/model/response/ResearchCallResponse.java.
type ResearchCallResponse struct {
	BasketId         int64      `json:"basketId"`
	LotSize          *string    `json:"lotSize"`
	SortOrder        *string    `json:"sortOrder"`
	Exchange         *string    `json:"exchange"`
	Token            *string    `json:"token"`
	TradingSymbol    *string    `json:"tradingSymbol"`
	Qty              *string    `json:"qty"`
	Price            *string    `json:"price"`
	Expiry           *Timestamp `json:"expiry"`
	Product          *string    `json:"product"`
	TransType        *string    `json:"transType"`
	PriceType        *string    `json:"priceType"`
	OrderType        *string    `json:"orderType"`
	Ret              *string    `json:"ret"`
	TriggerPrice     *string    `json:"triggerPrice"`
	DisClosedQty     *string    `json:"disClosedQty"`
	SpeclizationTag  *string    `json:"speclizationTag"`
	MktProtection    *string    `json:"mktProtection"`
	Target           *string    `json:"target"`
	Status           int        `json:"status"`
	StopLoss         *string    `json:"stopLoss"`
	TrailingStopLoss *string    `json:"trailingStopLoss"`
	FormattedInsName *string    `json:"formattedInsName"`
	WeekTag          *string    `json:"weekTag"`
	ValidityDays     *string    `json:"validityDays"`
	ExpiryDate       *string    `json:"expiryDate"`
	UserSegment      *string    `json:"userSegment"`
}

// ResearchReportDTO corresponds to in/codifi/basket/model/response/ResearchReportDTO.java.
type ResearchReportDTO struct {
	Id          int        `json:"id"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Attachment  *string    `json:"attachment"`
	Type        *string    `json:"type"`
	Url         *string    `json:"url"`
	CreatedOn   *Timestamp `json:"createdOn"`
}

// ScripDetailDto corresponds to in/codifi/basket/model/response/ScripDetailDto.java.
type ScripDetailDto struct {
	Token         *string `json:"token"`
	Exchange      *string `json:"exchange"`
	Qty           *string `json:"qty"`
	TradingSymbol *string `json:"tradingSymbol"`
}

// ScripDetailResponse corresponds to in/codifi/basket/model/response/ScripDetailResponse.java.
type ScripDetailResponse struct {
	Exchange         *string `json:"exchange"`
	Token            *string `json:"token"`
	Qty              *string `json:"qty"`
	HoldQty          *string `json:"holdQty"`
	TransType        *string `json:"transType"`
	Weightage        *string `json:"weightage"`
	TradingSymbol    *string `json:"tradingSymbol"`
	FormattedInsName *string `json:"formattedInsName"`
	Price            *string `json:"price"`
	Pdc              *string `json:"pdc"`
	Exposure         *string `json:"exposure"`
	OrderType        *string `json:"orderType"`
	PriceType        *string `json:"priceType"`
	MarketCap        *string `json:"marketCap"`
	Version          *int    `json:"version"`
}

// ScripHoldingModel corresponds to in/codifi/basket/model/response/ScripHoldingModel.java.
type ScripHoldingModel struct {
	Token              *string `json:"token"`
	Exchange           *string `json:"exchange"`
	TradingSymbol      *string `json:"tradingSymbol"`
	CurrentQty         int     `json:"currentQty"`
	RecommendedQty     int     `json:"recommendedQty"`
	RebalancedQty      int     `json:"rebalancedQty"`
	OrgRecoQty         int     `json:"orgRecoQty"`
	ExecutedPrice      float64 `json:"executedPrice"`
	ExecutedOn         *string `json:"executedOn"`
	Version            int     `json:"version"`
	RecommendedVersion int     `json:"recommendedVersion"`
	RebalancedAction   *string `json:"rebalancedAction"`
}

// SectorReportDetailsDTO corresponds to in/codifi/basket/model/response/SectorReportDetailsDTO.java.
type SectorReportDetailsDTO struct {
	Id           int        `json:"id"`
	Name         *string    `json:"name"`
	Description  *string    `json:"description"`
	Attachment   *string    `json:"attachment"`
	Url          *string    `json:"url"`
	Type         *string    `json:"type"`
	CreatedOn    *Timestamp `json:"createdOn"`
	CreatedBy    *int       `json:"createdBy"`
	UpdatedOn    *Timestamp `json:"updatedOn"`
	UpdatedBy    *int       `json:"updatedBy"`
	ActiveStatus int        `json:"activeStatus"`
}

// SectorReportResponse corresponds to in/codifi/basket/model/response/SectorReportResponse.java.
type SectorReportResponse struct {
	Id          *int64  `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Attachment  *string `json:"attachment"`
	Type        *string `json:"type"`
}

// SectorReportSummaryDTO corresponds to in/codifi/basket/model/response/SectorReportSummaryDTO.java.
type SectorReportSummaryDTO struct {
	Id          int        `json:"id"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Attachment  *string    `json:"attachment"`
	Type        *string    `json:"type"`
	Url         *string    `json:"url"`
	CreatedOn   *Timestamp `json:"createdOn"`
}

// SpanMarginResp corresponds to in/codifi/basket/model/response/SpanMarginResp.java.
type SpanMarginResp struct {
	Span *string `json:"span"`
}

// ThematicBasketDetailResponse corresponds to in/codifi/basket/model/response/ThematicBasketDetailResponse.java.
type ThematicBasketDetailResponse struct {
	BasketName          *string                `json:"basketName"`
	ShortDescription    *string                `json:"shortDescription"`
	LongDescription     *string                `json:"longDescription"`
	Tag                 *string                `json:"tag"`
	TagDescription      *string                `json:"tagDescription"`
	OverallWeightage    *string                `json:"overallWeightage"`
	Methodology         *string                `json:"methodology"`
	LaunchDate          *string                `json:"launchDate"`
	Rationale           *string                `json:"rationale"`
	RiskDescription     *string                `json:"riskDescription"`
	Analyst             *string                `json:"analyst"`
	Type                *string                `json:"type"`
	Category            *string                `json:"category"`
	Subcategory         *string                `json:"subcategory"`
	MinInvestmentAmount *string                `json:"minInvestmentAmount"`
	ExpectedReturn      *string                `json:"expectedReturn"`
	Risk                *string                `json:"risk"`
	Benchmark           *string                `json:"benchmark"`
	Exposure            *string                `json:"exposure"`
	InvestmentDuration  *string                `json:"investmentDuration"`
	ReviewFrequency     *string                `json:"reviewFrequency"`
	MinInvstAmt         *float64               `json:"minInvstAmt"`
	CurReturn           *string                `json:"curReturn"`
	CreatedOn           *Timestamp             `json:"createdOn"`
	ExpiryDate          *Timestamp             `json:"expiryDate"`
	Scrips              []ScripDetailResponse  `json:"scrips"`
	Documents           []SectorReportResponse `json:"documents"`
}

// ThematicBasketPreviewScripDTO corresponds to in/codifi/basket/model/response/ThematicBasketPreviewScripDTO.java.
type ThematicBasketPreviewScripDTO struct {
	FormattedInsName *string `json:"formattedInsName"`
	Token            *string `json:"token"`
	AdjWeightage     *string `json:"adjWeightage"`
	Qty              int     `json:"qty"`
	Exchange         *string `json:"exchange"`
	TransType        *string `json:"transType"`
	Weightage        *string `json:"weightage"`
	TradingSymbol    *string `json:"tradingSymbol"`
	Price            *string `json:"price"`
	Pdc              *string `json:"pdc"`
	OrderType        *string `json:"orderType"`
	PriceType        *string `json:"priceType"`
}

// ThematicBasketResponse corresponds to in/codifi/basket/model/response/ThematicBasketResponse.java.
type ThematicBasketResponse struct {
	Category            *string                       `json:"category"`
	SubCategory         *string                       `json:"subCategory"`
	BasketId            *string                       `json:"basketId"`
	BasketName          *string                       `json:"basketName"`
	ShortDescription    *string                       `json:"shortDescription"`
	LongDescription     *string                       `json:"longDescription"`
	Tag                 *string                       `json:"tag"`
	TotalInvstAmt       float64                       `json:"totalInvstAmt"`
	CurReturn           *string                       `json:"curReturn"`
	Rebalance           *string                       `json:"rebalance"`
	MinInvestmentAmount *string                       `json:"minInvestmentAmount"`
	CreatedOn           *Timestamp                    `json:"createdOn"`
	ExpiryDate          *Timestamp                    `json:"expiryDate"`
	Scrips              []ThematicBasketScripResponse `json:"scrips"`
}

// ThematicBasketScripResponse corresponds to in/codifi/basket/model/response/ThematicBasketScripResponse.java.
type ThematicBasketScripResponse struct {
	Exchange *string `json:"exchange"`
	Token    *string `json:"token"`
	Price    *string `json:"price"`
	Pdc      *string `json:"pdc"`
	Qty      *string `json:"qty"`
}

// ThematicMasterModel corresponds to in/codifi/basket/model/response/ThematicMasterModel.java.
type ThematicMasterModel struct {
	BasketId    int     `json:"basketId"`
	BasketName  *string `json:"basketName"`
	AnalystName *string `json:"analystName"`
}

// UserBasketHolding corresponds to in/codifi/basket/model/response/UserBasketHolding.java.
type UserBasketHolding struct {
	BasketName     *string          `json:"basketName"`
	UserId         *string          `json:"userId"`
	InvestedAmount float64          `json:"investedAmount"`
	CreatedDate    *string          `json:"createdDate"`
	ExecutedDate   *string          `json:"executedDate"`
	LotSize        *string          `json:"lotSize"`
	ScripList      []ScripDetailDto `json:"scripList"`
}

// UserHoldingInfo corresponds to in/codifi/basket/model/response/UserHoldingInfo.java.
type UserHoldingInfo struct {
	Token          *string    `json:"token"`
	Exchange       *string    `json:"exchange"`
	TradingSymbol  *string    `json:"tradingSymbol"`
	ExecutedQty    int        `json:"executedQty"`
	OrgRecoQty     int        `json:"orgRecoQty"`
	AvgPrice       float64    `json:"avgPrice"`
	LastExecutedOn *Timestamp `json:"lastExecutedOn"`
	Version        int        `json:"version"`
}

// BasketListModel corresponds to in/codifi/basket/ws/model/BasketListModel.java.
type BasketListModel struct {
	Exch   *string `json:"exch"`
	Qty    *string `json:"qty"`
	Symbol *string `json:"symbol"`
}

// BasketMarginRestReqModel corresponds to in/codifi/basket/ws/model/BasketMarginRestReqModel.java.
type BasketMarginRestReqModel struct {
	Symbol      *string           `json:"symbol"`
	Exch        *string           `json:"exch"`
	Qty         *string           `json:"qty"`
	Basketlists []BasketListModel `json:"basketlists"`
}

// BasketMarginRestRespModel corresponds to in/codifi/basket/ws/model/BasketMarginRestRespModel.java.
type BasketMarginRestRespModel struct {
	TotalRequirement   *string `json:"totalRequirement"`
	Stat               *string `json:"stat"`
	Emsg               *string `json:"emsg"`
	SpreadBenefit      *string `json:"spreadBenefit"`
	ExposureMarginPrst *string `json:"exposureMarginPrst"`
	SpanRequirement    *string `json:"spanRequirement"`
}

// CommonErrorModel corresponds to in/codifi/basket/ws/model/CommonErrorModel.java.
type CommonErrorModel struct {
	Emsg *string `json:"emsg"`
	Stat *string `json:"stat"`
	Text *string `json:"text"`
}

// OrderDetails corresponds to in/codifi/basket/ws/model/OrderDetails.java.
type OrderDetails struct {
	Exchange         *string `json:"exchange"`
	TradingSymbol    *string `json:"tradingSymbol"`
	Qty              *string `json:"qty"`
	Price            *string `json:"price"`
	OrderType        *string `json:"orderType"`
	Product          *string `json:"product"`
	PriceType        *string `json:"priceType"`
	TransType        *string `json:"transType"`
	Ret              *string `json:"ret"`
	TriggerPrice     *string `json:"triggerPrice"`
	DisclosedQty     *string `json:"disclosedQty"`
	MktProtection    *string `json:"mktProtection"`
	Target           *string `json:"target"`
	StopLoss         *string `json:"stopLoss"`
	TrailingStopLoss *string `json:"trailingStopLoss"`
	OrderNo          *string `json:"orderNo"`
	Source           *string `json:"source"`
	Token            *string `json:"token"`
	Remark           *string `json:"remark"`
}

// ResearchCallModelResponse corresponds to in/codifi/basket/ws/model/ResearchCallModelResponse.java.
type ResearchCallModelResponse struct {
	Id                    int        `json:"id"`
	SortOrder             int        `json:"sortOrder"`
	BasketName            *string    `json:"basketName"`
	UserId                *string    `json:"userId"`
	ExpiryDate            *Timestamp `json:"expiryDate"`
	CreatedOn             *string    `json:"createdOn"`
	CreatedBy             *string    `json:"createdBy"`
	SpeclizationTag       *string    `json:"speclizationTag"`
	Status                *string    `json:"status"`
	ShortDescription      *string    `json:"shortDescription"`
	LongDescription       *string    `json:"longDescription"`
	Remarks               *string    `json:"remarks"`
	Category              *string    `json:"category"`
	Channels              *string    `json:"channels"`
	SubCategory           *string    `json:"subCategory"`
	Tags                  *string    `json:"tags"`
	Researchcall          *string    `json:"researchcall"`
	VendorCode            *string    `json:"vendorCode"`
	IsExecuted            *string    `json:"isExecuted" gorm:"default:0"`
	IsVendorBasket        *string    `json:"isVendorBasket"`
	Source                *string    `json:"source"`
	ActiveStatus          *string    `json:"activeStatus"`
	SendPushNotification  *string    `json:"sendPushNotification"`
	PushNotificationTitle *string    `json:"pushNotificationTitle"`
	AnalystName           *string    `json:"analystName"`
	Attachement           *string    `json:"attachement"`
	InvestmentDuration    *string    `json:"investmentDuration"`
}

// SpanMarginPos corresponds to in/codifi/basket/ws/model/SpanMarginPos.java.
type SpanMarginPos struct {
	Prd      *string `json:"prd"`
	Exch     *string `json:"exch"`
	Instname *string `json:"instname"`
	Symname  *string `json:"symname"`
	Exd      *string `json:"exd"`
	Optt     *string `json:"optt"`
	Strprc   *string `json:"strprc"`
	Buyqty   *string `json:"buyqty"`
	Sellqty  *string `json:"sellqty"`
	Netqty   *string `json:"netqty"`
}

// SpanMarginRestReq corresponds to in/codifi/basket/ws/model/SpanMarginRestReq.java.
type SpanMarginRestReq struct {
	Actid *string         `json:"actid"`
	Pos   []SpanMarginPos `json:"pos"`
}

// SpanMarginRestResp corresponds to in/codifi/basket/ws/model/SpanMarginRestResp.java.
type SpanMarginRestResp struct {
	Stat       *string `json:"stat"`
	Emsg       *string `json:"emsg"`
	Span       *string `json:"span"`
	Expo       *string `json:"expo"`
	Span_trade *string `json:"span_trade"`
	Expo_trade *string `json:"expo_trade"`
}

// ClinetInfoModel corresponds to in/codifi/cache/model/ClinetInfoModel.java.
type ClinetInfoModel struct {
	UserId *string `json:"userId"`
	Ucc    *string `json:"ucc"`
	Name   *string `json:"name"`
	Email  *string `json:"email"`
}

// ContractMasterModel corresponds to in/codifi/cache/model/ContractMasterModel.java.
type ContractMasterModel struct {
	Exch             *string    `json:"exch"`
	Segment          *string    `json:"segment"`
	Token            *string    `json:"token"`
	AlterToken       *string    `json:"alterToken"`
	Symbol           *string    `json:"symbol"`
	TradingSymbol    *string    `json:"tradingSymbol"`
	FormattedInsName *string    `json:"formattedInsName"`
	Isin             *string    `json:"isin"`
	GroupName        *string    `json:"groupName"`
	InsType          *string    `json:"insType"`
	OptionType       *string    `json:"optionType"`
	StrikePrice      *string    `json:"strikePrice"`
	Expiry           *Timestamp `json:"expiry"`
	LotSize          *string    `json:"lotSize"`
	TickSize         *string    `json:"tickSize"`
	FreezQty         *string    `json:"freezQty"`
	Pdc              *string    `json:"pdc"`
	WeekTag          *string    `json:"weekTag"`
	CompanyName      *string    `json:"companyName"`
}

func Entities() []any {
	return []any{&BasketNameEntity{}, &BasketScripEntity{}, &DeviceMappingEntity{}, &OrderStatusFeedEntity{}, &ReasearchCallUsers{}, &ResearchCallStatusEntity{}, &ResearchcallOrderEntity{}, &ResearchcallScripEntity{}, &SectorReportsEntity{}, &ThematicExeMasterEntity{}, &ThematicExecDetails{}, &UserNotification{}, &VendorAppEntity{}}
}
