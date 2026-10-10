/*
  Warnings:

  - A unique constraint covering the columns `[userId]` on the table `Credits` will be added. If there are existing duplicate values, this will fail.

*/
-- CreateIndex
CREATE UNIQUE INDEX "Credits_userId_key" ON "Credits"("userId");
