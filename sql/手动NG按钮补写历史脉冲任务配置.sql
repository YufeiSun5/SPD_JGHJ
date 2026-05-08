-- ========================================================
-- 手动 NG 按钮补写历史脉冲任务配置
-- Manual NG button history-pulse backfill task config
-- 手動 NG ボタン履歴パルス補完タスク設定
-- 日期: 2026-05-08
--
-- CN:
--   前端生产计划页的 NG+1 / NG-1 按钮会手动触发 sys_tasks，
--   但不会经过 MQTT 采集层，因此不会自动写 sys_data_history。
--   本脚本只给四个 NG 手动按钮任务打开 write_manual_history 开关；
--   代码会在工单更新成功后，用任务自身 trigger_var_id 补写 val=1 历史脉冲。
-- EN:
--   Production-page NG+1 / NG-1 buttons manually trigger sys_tasks and bypass MQTT acquisition,
--   so sys_data_history is not written automatically. This script enables write_manual_history
--   only for the four manual NG button tasks; backend code backfills one val=1 pulse using each
--   task's trigger_var_id after the order update succeeds.
-- JP:
--   生産計画画面の NG+1 / NG-1 ボタンは sys_tasks を手動起動し MQTT 収集層を通らないため、
--   sys_data_history は自動で書かれません。本スクリプトは 4 つの手動 NG ボタンタスクだけに
--   write_manual_history を有効化し、工単更新成功後に trigger_var_id で val=1 履歴パルスを補完します。
--
-- 使用:
--   1. 当前代码已兼容旧任务配置：前端手动触发且 ng_qty_delta != 0 时，会按 trigger_var_id 自动补历史。
--   2. 本脚本用于把这个行为显式写入任务配置，便于后续维护和排查。
--   3. 先整段执行，查看第 2/4 步结果。
--   4. 确认四条 write_manual_history 都为 true 后，把最后一行 ROLLBACK 改为 COMMIT 再执行。
-- ========================================================

START TRANSACTION;

-- 1. 备份四个任务 / Backup the four tasks / 4 件のタスクをバックアップ
CREATE TABLE IF NOT EXISTS sys_tasks_bak_manual_history_20260508 LIKE sys_tasks;

INSERT INTO sys_tasks_bak_manual_history_20260508
SELECT *
FROM sys_tasks
WHERE task_id IN (25,31,32,33)
  AND NOT EXISTS (
    SELECT 1
    FROM sys_tasks_bak_manual_history_20260508 b
    WHERE b.task_id = sys_tasks.task_id
  );

-- 2. 更新前预览 / Preview before update / 更新前プレビュー
SELECT
  task_id,
  task_name,
  trigger_var_id,
  trigger_var_name,
  JSON_VALID(action_config) AS json_ok,
  JSON_EXTRACT(action_config, '$.operation') AS operation,
  JSON_EXTRACT(action_config, '$.op_params') AS op_params,
  JSON_EXTRACT(action_config, '$.op_params.write_manual_history') AS write_manual_history
FROM sys_tasks
WHERE task_id IN (25,31,32,33)
ORDER BY task_id;

-- 3. 打开手动历史脉冲补写开关 / Enable manual history pulse backfill / 手動履歴パルス補完を有効化
UPDATE sys_tasks
SET
  action_config = JSON_PRETTY(
    JSON_SET(
      CAST(action_config AS JSON),
      '$.op_params.write_manual_history',
      CAST('true' AS JSON)
    )
  ),
  updated_at = NOW()
WHERE task_id IN (25,31,32,33)
  AND JSON_VALID(action_config) = 1
  AND JSON_UNQUOTE(JSON_EXTRACT(action_config, '$.operation')) = 'increment_production_qty';

-- 4. 更新后复查 / Verify after update / 更新後確認
SELECT
  task_id,
  task_name,
  trigger_var_id,
  trigger_var_name,
  JSON_EXTRACT(action_config, '$.op_params') AS op_params,
  JSON_EXTRACT(action_config, '$.op_params.write_manual_history') AS write_manual_history
FROM sys_tasks
WHERE task_id IN (25,31,32,33)
ORDER BY task_id;

-- 5. 变更行数复查 / Row-count check / 行数確認
SELECT
  COUNT(*) AS enabled_manual_history_tasks
FROM sys_tasks
WHERE task_id IN (25,31,32,33)
  AND JSON_EXTRACT(action_config, '$.op_params.write_manual_history') = CAST('true' AS JSON);

-- 默认回滚，确认第 4/5 步正确后，把 ROLLBACK 改为 COMMIT。
-- Default is rollback. After confirming steps 4/5, change ROLLBACK to COMMIT.
-- 既定はロールバック。手順 4/5 を確認後、ROLLBACK を COMMIT に変更してください。
ROLLBACK;

-- 已提交后如需恢复 / Restore after committed if needed / COMMIT 後に戻す場合:
-- UPDATE sys_tasks t
-- JOIN sys_tasks_bak_manual_history_20260508 b ON b.task_id = t.task_id
-- SET
--   t.action_config = b.action_config,
--   t.updated_at = NOW()
-- WHERE t.task_id IN (25,31,32,33);
