/*
  Warnings:

  - The `status` column on the `JobRun` table would be dropped and recreated. This will lead to data loss if there is data in the column.

*/
-- CreateEnum
CREATE TYPE "JobRunStatus" AS ENUM ('QUEUED', 'RUNNING', 'SUCCEEDED', 'FAILED', 'INSUFFICIENT_CREDITS');

-- AlterTable
ALTER TABLE "JobRun" DROP COLUMN "status",
ADD COLUMN     "status" "JobRunStatus" NOT NULL DEFAULT 'QUEUED',
ALTER COLUMN "trace" SET DEFAULT '[]',
ALTER COLUMN "totalCost" SET DEFAULT 0,
ALTER COLUMN "totalTokens" SET DEFAULT 0,
ALTER COLUMN "startedAt" DROP NOT NULL,
ALTER COLUMN "finishedAt" DROP NOT NULL;
