CREATE TABLE `request_span` (
  `id` BIGINT NOT NULL,
  `annotation` VARCHAR(50) NULL COMMENT 'cs/sr/ss/cr',
  `trace_id` VARCHAR(100) NULL,
  `rpc_request_id` VARCHAR(100) NULL,
  `rpc_method_name` VARCHAR(100) NULL,
  `rpc_request_params` TEXT NULL,
  `rpc_response_error` TEXT NULL,
  `rpc_response_result` TEXT NULL,
  `target_server` TEXT NULL,
  `log_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `create_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`))
COMMENT = 'each request have some spans';

ALTER TABLE `request_span`
ADD INDEX `request_span_idx_trace_request` (`trace_id` ASC, `rpc_request_id` ASC);
;

ALTER TABLE `request_span`
ADD INDEX `request_span_idx_method_name` (`rpc_method_name` ASC);
;

CREATE TABLE `service_log` (
  `id` BIGINT NOT NULL,
  `service_name` VARCHAR(100) NOT NULL,
  `url` TEXT NOT NULL,
  `down_time` TIMESTAMP NULL,
  `create_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`));

ALTER TABLE `service_log`
ADD INDEX `service_log_idx_service_name` (`service_name` ASC);
;
