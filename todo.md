# 重构计划

## 文件表
字段调整
增加过期时间，结合用户自己配置的回收站配置

## file_store表重构完成

## user表重构完成

## file_folder表不需要重构

## file表有个字段filePath待定

## service/admin完成

## service/filestore完成

## service/filefloder完成

## file
- 文件删除时，进入软删，deleted_at到期后回收站彻底删，然后触发异步cos删
- cos上传验证
- 分片需要支持（前端未接入）
- 标签不创联了

## 回收站配置修改和查（前端也未接入）

## tag增删改查（前端也没接入）

## 分享加个审核（然后也需要验证稳定性）

## 前端没有接入秒传+分片传+续传模块
