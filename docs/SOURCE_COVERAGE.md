# Java source coverage

This index accounts for source files. It is not a proof of identical behavior; see PARITY.md for limitations and changes.

| Java source | Go counterpart |
|---|---|
| `com/sas/cache/HazleCacheController.java` | internal/app/cache.go |
| `com/sas/dto/CustomerDTO.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `com/sas/dto/DefaultLoginDTO.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/config/ApplicationProperties.java` | internal/config/config.go; internal/app/{app,cache,logging,notifications}.go |
| `in/codifi/basket/config/ClickHouseLoggerProducer.java` | internal/config/config.go; internal/app/{app,cache,logging,notifications}.go |
| `in/codifi/basket/config/ConcorentHashMapConfig.java` | internal/config/config.go; internal/app/{app,cache,logging,notifications}.go |
| `in/codifi/basket/config/FirebaseConfig.java` | internal/config/config.go; internal/app/{app,cache,logging,notifications}.go |
| `in/codifi/basket/config/HazelcastConfig.java` | internal/config/config.go; internal/app/{app,cache,logging,notifications}.go |
| `in/codifi/basket/config/RestServiceProperties.java` | internal/config/config.go; internal/app/{app,cache,logging,notifications}.go |
| `in/codifi/basket/controller/BasketOrderApiController.java` | internal/app/app.go route table and corresponding module handler |
| `in/codifi/basket/controller/BasketOrderController.java` | internal/app/app.go route table and corresponding module handler |
| `in/codifi/basket/controller/DefaultRestController.java` | internal/app/auth.go |
| `in/codifi/basket/controller/HoldingsController.java` | internal/app/holdings.go |
| `in/codifi/basket/controller/IResearchCallApiController.java` | internal/app/app.go route table and corresponding module handler |
| `in/codifi/basket/controller/ThematicBasketController.java` | internal/app/app.go route table and corresponding module handler |
| `in/codifi/basket/controller/spec/HoldingsControllerSpec.java` | internal/app/app.go route table and corresponding module handler |
| `in/codifi/basket/controller/spec/IBasketOrderApiController.java` | internal/app/app.go route table and corresponding module handler |
| `in/codifi/basket/controller/spec/IBasketOrderController.java` | internal/app/app.go route table and corresponding module handler |
| `in/codifi/basket/controller/spec/ICacheController.java` | internal/app/app.go route table and corresponding module handler |
| `in/codifi/basket/controller/spec/IResearchCallApiControllerSpec.java` | internal/app/app.go route table and corresponding module handler |
| `in/codifi/basket/controller/spec/ThematicBasketControllerSpec.java` | internal/app/app.go route table and corresponding module handler |
| `in/codifi/basket/dao/HoldingsDao.java` | internal/app/holdings.go |
| `in/codifi/basket/dao/ResearchCallDAO.java` | internal/app/research.go; internal/app/repository.go |
| `in/codifi/basket/entity/logs/AccessLogModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/logs/RestAccessLogModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/BasketNameEntity.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/BasketScripEntity.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/CommonEntity.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/DeviceMappingEntity.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/OrderStatusFeedEntity.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/ReasearchCallUsers.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/ResearchCallStatusEntity.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/ResearchcallOrderEntity.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/ResearchcallScripEntity.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/SectorReportsEntity.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/ThematicExeMasterEntity.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/ThematicExecDetails.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/UserNotification.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/entity/primary/VendorAppEntity.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/filter/AccessLogFilter.java` | internal/app/logging.go |
| `in/codifi/basket/filter/BaskerOrderFilter.java` | internal/app/app.go |
| `in/codifi/basket/filter/TokenResource.java` | internal/app/app.go (see PARITY.md) |
| `in/codifi/basket/loader/InitialLoader.java` | internal/app/app.go |
| `in/codifi/basket/model/request/AdminBasketOrderReq.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/BasketMarginRequest.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/BasketOrderListUpdateReq.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/BasketOrderReq.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/ExecuteBasketOrderReq.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/OrderBookReqModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/ResearchCallModelRequest.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/ResearchCallRequest.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/RetrieveBasketModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/ScripRequest.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/ScripRequestModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/SendNoficationReqModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/SpanMarginReq.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/ThematicBasketMaster.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/ThematicBasketRequest.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/request/ThematicBasketScrip.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/AdminRecommendation.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/BasketHoldingsResponse.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/BasketMasterModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/BasketScripModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/GenericOrderBookResp.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/GenericOrderResp.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/GenericResponse.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/ResearchCallResponse.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/ResearchReportDTO.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/ScripDetailDto.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/ScripDetailResponse.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/ScripHoldingModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/SectorReportDetailsDTO.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/SectorReportResponse.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/SectorReportSummaryDTO.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/SpanMarginResp.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/ThematicBasketDetailResponse.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/ThematicBasketPreviewScripDTO.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/ThematicBasketResponse.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/ThematicBasketScripResponse.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/ThematicMasterModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/UserBasketHolding.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/model/response/UserHoldingInfo.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/repository/AccessLogManager.java` | internal/app/logging.go |
| `in/codifi/basket/repository/BasketOrderEntityManager.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/repository/BasketOrderRepository.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/repository/BasketScripRepository.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/repository/DeviceMappingRepository.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/repository/OrderStatusFeedDAO.java` | internal/app/orderbook.go |
| `in/codifi/basket/repository/ResearchOrderRepository.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/repository/ResearchcallScripRepository.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/repository/ResearchcallUserRepository.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/repository/SectorReportsRepository.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/repository/ThematicDao.java` | internal/app/thematic.go; internal/app/repository.go |
| `in/codifi/basket/repository/ThematicExeDetailRepository.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/repository/ThematicExeMasterEntityRepository.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/repository/UserNotificationRepository.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/repository/VendorAppRepository.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/repository/researchcallStatusRepository.java` | GORM models/queries in corresponding module service; internal/app/repository.go |
| `in/codifi/basket/scheduler/Scheduler.java` | internal/app/app.go; internal/app/admin.go |
| `in/codifi/basket/service/BasketOrderApiService.java` | internal/app/admin.go |
| `in/codifi/basket/service/BasketOrderService.java` | internal/app/basket.go; internal/app/upstream.go |
| `in/codifi/basket/service/IResearchCallApiService.java` | internal/app/research.go |
| `in/codifi/basket/service/ThematicBasketService.java` | internal/app/thematic.go; internal/app/orderbook.go |
| `in/codifi/basket/service/UserNotificationService.java` | internal/app/notifications.go |
| `in/codifi/basket/service/spec/IBasketOrderApiService.java` | Go module handler/service functions; no framework proxy required |
| `in/codifi/basket/service/spec/IBasketOrderService.java` | Go module handler/service functions; no framework proxy required |
| `in/codifi/basket/service/spec/IResearchCallApiServiceSpec.java` | Go module handler/service functions; no framework proxy required |
| `in/codifi/basket/service/spec/InterfaceCacheService.java` | Go module handler/service functions; no framework proxy required |
| `in/codifi/basket/service/spec/ThematicBasketServiceSpec.java` | Go module handler/service functions; no framework proxy required |
| `in/codifi/basket/service/spec/UserNotificationSpec.java` | Go module handler/service functions; no framework proxy required |
| `in/codifi/basket/utility/AppConstants.java` | internal/config/config.go; internal/app/common.go (protocol messages) |
| `in/codifi/basket/utility/AppUtil.java` | internal/app/auth.go; internal/app/cache.go |
| `in/codifi/basket/utility/CodifiUtil.java` | internal/utility/legacy.go; upstream TLS config |
| `in/codifi/basket/utility/CommonUtils.java` | internal/utility/legacy.go; upstream TLS config |
| `in/codifi/basket/utility/FcmNotificationUtils.java` | internal/app/notifications.go |
| `in/codifi/basket/utility/PrepareResponse.java` | internal/app/common.go |
| `in/codifi/basket/utility/PushNoficationUtils.java` | internal/app/notifications.go |
| `in/codifi/basket/utility/StringUtil.java` | internal/utility/legacy.go; standard strings package |
| `in/codifi/basket/utility/Test.java` | test fixtures only; hardcoded production-device sender intentionally not ported |
| `in/codifi/basket/utility/ValidateUtil.java` | internal/utility/legacy.go |
| `in/codifi/basket/ws/model/BasketListModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/ws/model/BasketMarginRestReqModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/ws/model/BasketMarginRestRespModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/ws/model/CommonErrorModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/ws/model/OrderDetails.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/ws/model/ResearchCallModelResponse.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/ws/model/SpanMarginPos.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/ws/model/SpanMarginRestReq.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/ws/model/SpanMarginRestResp.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/basket/ws/service/InternalRestService.java` | internal/app/upstream.go |
| `in/codifi/basket/ws/service/OrdersRestService.java` | internal/app/orderbook.go |
| `in/codifi/basket/ws/service/SpanMarginRestService.java` | internal/app/upstream.go |
| `in/codifi/basket/ws/service/spec/InternalRestServiceSpec.java` | Go module handler/service functions; no framework proxy required |
| `in/codifi/cache/AccessLogCache.java` | internal/app/logging.go (synchronous bounded writes) |
| `in/codifi/cache/CacheController.java` | internal/app/app.go route table and corresponding module handler |
| `in/codifi/cache/CacheService.java` | internal/app/admin.go |
| `in/codifi/cache/model/ClinetInfoModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
| `in/codifi/cache/model/ContractMasterModel.java` | internal/model/java_models.go (and JDBC extensions in tables.go) |
