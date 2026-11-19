<template>
  <div class="space-y-6 max-w-6xl">
    <!-- Main Connections Card -->
    <UCard class="border border-gray-200 dark:border-slate-800 dark:bg-slate-900 dark:ring-slate-800 rounded-2xl shadow-md">
      <template #header>
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('settings.dataSource.title') }}</h3>
            <p class="text-md text-gray-500 dark:text-gray-400 mt-0.5">
              {{ t('settings.dataSource.subtitle') }}
            </p>
          </div>
          <div class="flex items-center gap-2 flex-wrap self-start sm:self-auto shrink-0">
            <SourceSelectorDropdown />
            <UButton
              v-if="connections.length > 1"
              color="secondary"
              variant="subtle"
              icon="i-lucide-refresh-cw"
              class="rounded-xl font-bold"
              :loading="isBatchSyncing"
              :disabled="isBatchSyncing || getSyncCooldownRemaining('__all__') > 0"
              @click="syncAllSources"
            >
              {{ t('settings.dataSource.syncAll') }}<template v-if="getSyncCooldownRemaining('__all__') > 0"> ({{ getSyncCooldownRemaining('__all__') }}s)</template>
            </UButton>
            <UButton
              color="primary"
              icon="i-lucide-plus"
              class="rounded-xl font-bold"
              @click="openAddModal"
            >
              {{ t('settings.dataSource.addConnection') }}
            </UButton>
          </div>
        </div>
      </template>

      <!-- Filter Controls -->
      <div class="flex flex-col sm:flex-row items-center justify-between gap-3 mb-6">
        <div class="w-full sm:w-72">
          <CommonSearchInput
            v-model="searchQuery"
            :placeholder="t('settings.dataSource.searchPlaceholder')"
            size="md"
            class="w-full"
          />
        </div>
        <div class="flex items-center gap-2 w-full sm:w-auto justify-end">
          <USelect
            v-model="statusFilter"
            :items="statusOptions"
            size="md"
            class="w-36"
          />
          <USelect
            v-model="typeFilter"
            :items="typeOptions"
            size="md"
            class="w-40"
          />
        </div>
      </div>

      <!-- Connections Grid List -->
      <div class="space-y-4">
        <div
          v-for="conn in filteredConnections"
          :key="conn.id"
          class="p-5 rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 hover:border-primary-500/50 transition-all duration-200 shadow-md"
        >
          <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
            <!-- Left Info -->
            <div class="flex items-start gap-4 min-w-0">
              <div
                class="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 border"
                :class="getProviderColorClass(conn.type)"
              >
                <UIcon :name="getProviderIcon(conn.type)" class="w-6 h-6" />
              </div>

              <div class="space-y-1 min-w-0 flex-1">
                <div class="flex items-center gap-2.5 flex-wrap">
                  <h4 class="font-bold text-gray-900 dark:text-white text-base truncate">{{ conn.name }}</h4>
                  <UBadge :color="getStatusBadgeColor(conn.status)" variant="subtle" size="sm" class="rounded-lg font-semibold flex items-center gap-1">
                    <span
                      class="w-1.5 h-1.5 rounded-full"
                      :class="getStatusDotClass(conn.status)"
                    />
                    {{ conn.status }}
                  </UBadge>
                  <UBadge color="neutral" variant="outline" size="sm" class="rounded-lg font-mono text-md">
                    {{ conn.environment }}
                  </UBadge>
                </div>

                <div class="flex items-center gap-x-4 gap-y-1 text-md text-gray-500 dark:text-gray-400 flex-wrap">
                  <span class="flex items-center gap-1 font-mono">
                    <UIcon name="i-lucide-server" class="w-3.5 h-3.5" />
                    {{ conn.host }}:{{ conn.port }}
                  </span>
                  <span class="flex items-center gap-1">
                    <UIcon name="i-lucide-database" class="w-3.5 h-3.5" />
                    {{ conn.database }}
                  </span>
                  <span class="flex items-center gap-1">
                    <UIcon name="i-lucide-refresh-cw" class="w-3.5 h-3.5" />
                    {{ normalizeSyncSchedule(conn.syncSchedule) }}
                  </span>
                  <span
                    v-if="conn.nextSyncAt && normalizeSyncSchedule(conn.syncSchedule) !== 'Manual Only'"
                    class="flex items-center gap-1 text-primary-600 dark:text-primary-400 font-medium"
                  >
                    <UIcon name="i-lucide-clock" class="w-3.5 h-3.5" />
                    {{ t('settings.dataSource.nextSync') }}: {{ formatSyncDateTime(conn.nextSyncAt) }}
                  </span>
                </div>

                <p v-if="conn.lastError" class="text-md text-rose-600 dark:text-rose-400 font-medium pt-1 flex items-center gap-1">
                  <UIcon name="i-lucide-alert-circle" class="w-3.5 h-3.5 shrink-0" />
                  {{ conn.lastError }}
                </p>
              </div>
            </div>

            <!-- Right Meta & Actions -->
            <div class="flex items-center justify-between lg:justify-end gap-3 pt-3 lg:pt-0 border-t lg:border-t-0 border-gray-100 dark:border-gray-800">
              <div class="text-left lg:text-right text-md text-gray-500 dark:text-gray-400">
                <p class="font-medium text-gray-700 dark:text-gray-300">{{ t('settings.dataSource.lastSynced') }}</p>
                <p class="font-mono text-md">{{ formatConnLastSync(conn) }}</p>
              </div>

              <div class="flex items-center gap-2 flex-wrap sm:flex-nowrap">
                <UButton
                  color="neutral"
                  variant="subtle"
                  size="md"
                  icon="i-lucide-table-properties"
                  class="rounded-xl font-semibold"
                  @click="openSchemaExplorer(conn)"
                >
                  {{ t('settings.dataSource.schemaExplorer') }}
                </UButton>

                <UButton
                  v-if="getActiveMappingsCountForConn(conn) > 0"
                  color="success"
                  variant="subtle"
                  size="md"
                  icon="i-lucide-download"
                  class="rounded-xl font-semibold"
                  :loading="importingConnId === conn.id && isImportingMappings"
                  :disabled="isImportingMappings"
                  @click="importAllMappingsForConn(conn)"
                >
                  Impor Semua ({{ getActiveMappingsCountForConn(conn) }})
                </UButton>

                <UButton
                  color="neutral"
                  variant="subtle"
                  size="md"
                  icon="i-lucide-plug"
                  class="rounded-xl font-semibold"
                  :loading="testingId === conn.id"
                  @click="testConnection(conn)"
                >
                  {{ t('settings.dataSource.test') }}
                </UButton>

                <UButton
                  color="primary"
                  variant="subtle"
                  size="md"
                  icon="i-lucide-refresh-cw"
                  class="rounded-xl font-semibold"
                  :loading="syncingId === conn.id"
                  :disabled="syncingId === conn.id || getSyncCooldownRemaining(conn.id) > 0"
                  @click="triggerSync(conn)"
                >
                  {{ t('settings.dataSource.syncNow') }}<template v-if="getSyncCooldownRemaining(conn.id) > 0"> ({{ getSyncCooldownRemaining(conn.id) }}s)</template>
                </UButton>

                <UDropdownMenu :items="getConnMenuItems(conn)">
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="md"
                    icon="i-lucide-more-vertical"
                    class="rounded-xl"
                  />
                </UDropdownMenu>
              </div>
            </div>
          </div>
        </div>

        <div v-if="filteredConnections.length === 0" class="text-center py-12 border-2 border-dashed border-gray-200 dark:border-gray-800 rounded-2xl">
          <UIcon name="i-lucide-database-zap" class="w-12 h-12 text-gray-400 dark:text-gray-600 mx-auto mb-3" />
          <h4 class="text-base font-bold text-gray-900 dark:text-white">{{ t('settings.dataSource.noSources') }}</h4>
          <p class="text-md text-gray-500 dark:text-gray-400 max-w-sm mx-auto mt-1 mb-4">
            {{ t('settings.dataSource.noSourcesSub') }}
          </p>
          <UButton color="primary" size="sm" icon="i-lucide-plus" class="rounded-xl font-bold" @click="openAddModal">
            {{ t('settings.dataSource.addConnection') }}
          </UButton>
        </div>
      </div>
    </UCard>

    <!-- Bulk Import Result Summary Card (when triggered from Connection List) -->
    <UCard
      v-if="showImportSummaryPanel && lastBulkImportSummary && !isSchemaModalOpen"
      class="bg-white dark:bg-slate-900 rounded-xl border border-emerald-200 dark:border-emerald-900/50 shadow-sm"
    >
      <template #header>
        <div class="flex items-center justify-between gap-3">
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-check-circle-2" class="w-5 h-5 text-emerald-500" />
            <h3 class="text-base font-bold text-gray-900 dark:text-white">
              Ringkasan Impor Mapping &mdash; {{ lastBulkImportSummary.connectionName }}
            </h3>
            <UBadge color="success" variant="subtle" size="sm" class="font-mono">
              +{{ lastBulkImportSummary.totalRecordsAdded }} data baru
            </UBadge>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            size="xs"
            icon="i-lucide-x"
            @click="showImportSummaryPanel = false"
          />
        </div>
      </template>

      <div class="space-y-2.5">
        <div
          v-for="res in lastBulkImportSummary.results"
          :key="res.key"
          class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 p-3 rounded-xl border text-xs"
          :class="res.status === 'imported'
            ? 'bg-emerald-50/50 dark:bg-emerald-950/20 border-emerald-200 dark:border-emerald-900/50'
            : res.status === 'manual_required'
              ? 'bg-amber-50/50 dark:bg-amber-950/20 border-amber-200 dark:border-amber-900/50'
              : res.status === 'skipped_duplicate'
                ? 'bg-gray-50 dark:bg-gray-800/40 border-gray-200 dark:border-gray-700'
                : 'bg-rose-50/50 dark:bg-rose-950/20 border-rose-200 dark:border-rose-900/50'"
        >
          <div class="space-y-0.5 min-w-0">
            <div class="flex items-center gap-2 flex-wrap">
              <span class="font-mono font-bold text-gray-900 dark:text-white">{{ res.tableName }}</span>
              <span class="text-gray-400">→</span>
              <span class="font-semibold text-gray-800 dark:text-gray-200">{{ res.targetModule }}</span>
              <UBadge
                :color="res.status === 'imported' ? 'success' : res.status === 'manual_required' ? 'warning' : res.status === 'skipped_duplicate' ? 'neutral' : 'error'"
                variant="subtle"
                size="xs"
              >
                {{ res.status === 'imported' ? `+${res.addedCount} Diimpor` : res.status === 'manual_required' ? 'Konfirmasi Manual' : res.status === 'skipped_duplicate' ? 'Sudah Sinkron' : res.status === 'incompatible' ? 'Tidak Cocok' : 'Gagal' }}
              </UBadge>
            </div>
            <p class="text-gray-600 dark:text-gray-400">{{ res.message }}</p>
          </div>

          <UButton
            v-if="res.route && res.route !== '/' && res.status !== 'incompatible'"
            :to="res.route"
            color="primary"
            variant="subtle"
            size="xs"
            icon="i-lucide-external-link"
            class="rounded-lg font-semibold shrink-0"
          >
            {{ res.status === 'manual_required' ? 'Buka & Konfirmasi' : 'Lihat Fitur' }}
          </UButton>
        </div>
      </div>
    </UCard>

    <!-- Recent Data Integration Activity Card -->
    <UCard class="bg-white dark:bg-slate-900 rounded-xl border border-gray-200 dark:border-slate-800 shadow-sm ring-0 divide-gray-200 dark:divide-slate-800">
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('settings.dataSource.recentLogsTitle') }}</h3>
          <UBadge color="neutral" variant="subtle" size="sm" class="rounded-lg font-medium">{{ t('settings.dataSource.realtimeTelemetry') }}</UBadge>
        </div>
      </template>

      <div class="overflow-x-auto">
        <table class="w-full text-md text-left">
          <thead class="bg-gray-100 dark:bg-gray-800/50 text-gray-600 dark:text-gray-400 font-semibold border-b border-gray-200 dark:border-gray-800">
            <tr>
              <th class="py-3 px-4">{{ t('settings.dataSource.thName') }}</th>
              <th class="py-3 px-4">{{ t('settings.dataSource.thEventType') }}</th>
              <th class="py-3 px-4">{{ t('settings.dataSource.thStatus') }}</th>
              <th class="py-3 px-4">{{ t('settings.dataSource.thRecords') }}</th>
              <th class="py-3 px-4">{{ t('settings.dataSource.thDuration') }}</th>
              <th class="py-3 px-4 text-right">{{ t('settings.dataSource.thTimestamp') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800/60 font-medium">
            <tr v-for="log in activityLogs" :key="log.id" class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
              <td class="py-3 px-4 text-gray-900 dark:text-white font-bold flex items-center gap-2">
                <UIcon :name="getProviderIcon(log.type)" class="w-4 h-4 text-primary-500" />
                {{ log.connName }}
              </td>
              <td class="py-3 px-4 text-gray-600 dark:text-white font-mono">{{ log.event }}</td>
              <td class="py-3 px-4">
                <UBadge :color="log.status === 'SUCCESS' ? 'success' : log.status === 'RUNNING' ? 'info' : 'error'" variant="subtle" size="md" class="rounded-md font-semibold">
                  {{ log.status }}
                </UBadge>
              </td>
              <td class="py-3 px-4 font-mono text-gray-700 dark:text-gray-300">{{ log.records.toLocaleString() }}</td>
              <td class="py-3 px-4 font-mono text-gray-500 dark:text-gray-400">{{ log.duration }}</td>
              <td class="py-3 px-4 text-right font-mono text-gray-500 dark:text-gray-400">{{ log.timestamp }}</td>
            </tr>
            <tr v-if="activityLogs.length === 0">
              <td colspan="6" class="py-8 text-center text-gray-500 dark:text-gray-400 text-md">
                {{ t('settings.dataSource.noLogs') }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </UCard>

    <!-- Add / Edit Modal -->
    <UModal
      v-model:open="isModalOpen"
      :title="isEditMode ? t('settings.dataSource.modalEditTitle') : t('settings.dataSource.modalAddTitle')"
      :description="isEditMode ? t('settings.dataSource.modalEditDesc') : t('settings.dataSource.modalAddDesc')"
      :close="{ color: 'neutral', variant: 'outline', onClick: closeModal }"
      dismissible
      class="w-[calc(100vw-2rem)] sm:w-full sm:max-w-2xl"
      :ui="{
        content: 'bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        header: 'px-6 py-5 border-b border-gray-100 dark:border-gray-800 pr-14',
        title: 'text-lg font-bold text-gray-900 dark:text-white leading-snug',
        description: 'text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-1.5 leading-relaxed',
        body: 'p-4 sm:p-6 space-y-4 bg-white dark:bg-gray-900 max-h-[80vh] overflow-y-auto',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
    >
      <template #body>
        <form @submit.prevent="saveConnection" class="space-y-4">
          <!-- PostgreSQL Option Quick Info Banner -->
          <div class="p-3.5 rounded-xl bg-primary-300 border-2 border-secondary-500 text-md text-secondary-900 flex items-start gap-2.5 shadow-sm">
            <UIcon name="i-lucide-zap" class="w-4 h-4 text-secondary-700 shrink-0 mt-0.5" />
            <div>
              <span class="font-bold text-secondary-900">{{ t('settings.dataSource.bannerTitle')  }} </span>
              <span>&nbsp;</span>
              <span class="text-secondary-900"> {{ t('settings.dataSource.bannerDesc') }}</span>
            </div>
          </div>

          <!-- Provider Type Selection grid -->
          <div>
            <label class="block text-md font-bold text-gray-700 dark:text-gray-200 mb-2">{{ t('settings.dataSource.selectProvider') }}</label>
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
              <button
                v-for="provider in providers"
                :key="provider.id"
                type="button"
                class="relative p-3 rounded-xl text-left flex flex-col items-center justify-center gap-1.5 transition-all duration-150"
                :class="form.type === provider.id
                  ? 'border-2 border-secondary-500 bg-primary-300 text-secondary-900 font-bold shadow-md'
                  : 'border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/40 hover:border-gray-300 dark:hover:border-gray-700 text-gray-700 dark:text-gray-300'"
                @click="selectProvider(provider)"
              >
                <span
                  v-if="provider.badge"
                  class="absolute top-18 px-1.5 py-0.5 rounded-full text-[9px] font-extrabold uppercase tracking-wider bg-emerald-700/15 text-emerald-600 dark:bg-emerald-900 dark:text-white border border-emerald-500/30"
                >
                  {{ provider.badge }}
                </span>
                <UIcon :name="provider.icon" class="w-6 h-6 mt-1" />
                <span class="text-md font-semibold">{{ provider.name }}</span>
              </button>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-2">
            <UFormField :label="t('settings.dataSource.connName')" required :help="t('settings.dataSource.connNameHelp')">
              <UInput v-model="form.name" :placeholder="t('settings.dataSource.connNamePlaceholder')" size="md" class="w-full" />
            </UFormField>

            <UFormField :label="t('settings.dataSource.environment')" required>
              <USelect v-model="form.environment" :items="['Production', 'Staging', 'Development']" size="md" class="w-full" />
            </UFormField>

            <UFormField :label="t('settings.dataSource.host')" required>
              <UInput v-model="form.host" placeholder="localhost, postgres-cbs-sim, or db.internal.net" size="md" class="w-full font-mono text-md" />
            </UFormField>

            <UFormField :label="t('settings.dataSource.port')" required>
              <UInput v-model.number="form.port" type="number" placeholder="5432" size="md" class="w-full font-mono text-md" />
            </UFormField>

            <UFormField :label="t('settings.dataSource.database')" required>
              <UInput v-model="form.database" placeholder="cbs_simulator or core_banking_db" size="md" class="w-full font-mono text-md" />
            </UFormField>

            <UFormField :label="t('settings.dataSource.syncFrequency')">
              <USelect
                v-model="form.syncSchedule"
                :items="[...SYNC_FREQUENCY_OPTIONS]"
                size="md"
                class="w-full"
              />
            </UFormField>

            <UFormField :label="t('settings.dataSource.username')">
              <UInput v-model="form.username" placeholder="audit_reader" size="md" class="w-full font-mono text-md" />
            </UFormField>

            <UFormField :label="t('settings.dataSource.password')">
              <UInput
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="••••••••••••"
                size="md"
                class="w-full font-mono text-md"
              >
                <template #trailing>
                  <UButton
                    color="neutral"
                    variant="ghost"
                    size="md"
                    :icon="showPassword ? 'i-lucide-eye-off' : 'i-lucide-eye'"
                    @click="() => { showPassword = !showPassword }"
                  />
                </template>
              </UInput>
            </UFormField>
          </div>

          <!-- Feature Scope Selection -->
          <div class="p-3.5 rounded-xl bg-gray-50 dark:bg-gray-800/40 border border-gray-200 dark:border-gray-800 space-y-2 mt-2">
            <div class="flex items-center justify-between">
              <label class="text-md font-bold text-gray-900 dark:text-white flex items-center gap-1.5">
                <UIcon name="i-lucide-layers" class="w-4 h-4 text-primary-500" />
                {{ t('settings.dataSource.enabledScopes') }}
              </label>
              <span class="text-[11px] text-gray-500 dark:text-gray-400">{{ t('settings.dataSource.enabledScopesSub') }}</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2.5 pt-1">
              <label
                v-for="scope in availableScopes"
                :key="scope.id"
                class="flex flex-col p-2.5 rounded-xl text-md cursor-pointer transition-colors select-none"
                :class="form.scopes.includes(scope.id)
                  ? 'border-2 border-secondary-500 bg-primary-300 text-secondary-900 font-semibold shadow-sm'
                  : 'border border-gray-200 dark:border-gray-700/60 text-gray-600 dark:text-gray-400 hover:border-gray-300 dark:hover:border-gray-600'"
              >
                <div class="flex items-center gap-2">
                  <UCheckbox
                    :model-value="form.scopes.includes(scope.id)"
                    color="secondary"
                    size="sm"
                    @update:model-value="toggleScope(scope.id)"
                  />
                  <span class="font-bold" :class="form.scopes.includes(scope.id) ? 'text-secondary-900' : 'text-gray-900 dark:text-white'">{{ scope.label }}</span>
                </div>
                <span class="text-[10px] mt-1 pl-6" :class="form.scopes.includes(scope.id) ? 'text-secondary-800' : 'text-gray-500 dark:text-gray-400'">{{ scope.description }}</span>
              </label>
            </div>
          </div>

          <div class="flex items-center justify-between p-3.5 rounded-xl bg-gray-50 dark:bg-gray-800/60 border border-gray-200 dark:border-gray-800 mt-2">
            <div class="space-y-0.5">
              <p class="text-md font-bold text-gray-900 dark:text-white">{{ t('settings.dataSource.enableSsl') }}</p>
              <p class="text-[11px] text-gray-500 dark:text-gray-400">{{ t('settings.dataSource.enableSslSub') }}</p>
            </div>
            <USwitch v-model="form.ssl" color="primary" />
          </div>

          <!-- Test result banner inside modal -->
          <div
            v-if="modalTestResult"
            class="p-3 rounded-xl text-md font-medium flex items-center gap-2 border"
            :class="modalTestResult.success
              ? 'bg-emerald-500/10 dark:bg-emerald-950/40 border-emerald-500/20 dark:border-emerald-900/50 text-emerald-600 dark:text-emerald-400'
              : 'bg-rose-500/10 dark:bg-rose-950/40 border-rose-500/20 dark:border-rose-900/50 text-rose-600 dark:text-rose-400'"
          >
            <UIcon :name="modalTestResult.success ? 'i-lucide-check-circle-2' : 'i-lucide-x-circle'" class="w-4 h-4 shrink-0" />
            <span>{{ modalTestResult.message }}</span>
          </div>

          <div class="flex items-center justify-between pt-4 border-t border-gray-200 dark:border-gray-800">
            <UButton
              type="button"
              color="neutral"
              variant="subtle"
              icon="i-lucide-plug"
              class="rounded-xl font-semibold text-md"
              :loading="isModalTesting"
              @click="testModalConnection"
            >
              {{ t('settings.dataSource.testConn') }}
            </UButton>

            <div class="flex items-center gap-2">
              <UButton type="button" color="neutral" variant="subtle" class="rounded-xl font-semibold" @click="closeModal">
                {{ t('settings.dataSource.cancel') }}
              </UButton>
              <UButton type="submit" color="primary" class="rounded-xl font-bold" :loading="isSaving">
                {{ isEditMode ? t('settings.dataSource.updateConn') : t('settings.dataSource.createConn') }}
              </UButton>
            </div>
          </div>
        </form>
      </template>
    </UModal>

    <!-- Schema Introspection & Data Mapping Modal -->
    <UModal
      v-model:open="isSchemaModalOpen"
      :title="t('settings.dataSource.schemaTitle')"
      :description="t('settings.dataSource.schemaSubtitle')"
      :close="{ color: 'neutral', variant: 'outline', onClick: closeSchemaModal }"
      dismissible
      class="w-[calc(100vw-2rem)] sm:w-full sm:max-w-5xl"
      :ui="{
        content: 'bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-2xl overflow-hidden',
        header: 'px-6 py-5 border-b border-gray-100 dark:border-gray-800 pr-14',
        title: 'text-lg font-bold text-gray-900 dark:text-white leading-snug',
        description: 'text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-1.5 leading-relaxed',
        body: 'p-4 sm:p-6 space-y-6 bg-white dark:bg-gray-900 max-h-[80vh] overflow-y-auto',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
    >
      <template #body>
        <div v-if="selectedConnForSchema" class="space-y-6">
          <!-- Connection Summary Header Badge -->
          <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-800">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-blue-500/10 text-blue-500 flex items-center justify-center font-bold">
                <UIcon :name="getProviderIcon(selectedConnForSchema.type)" class="w-5 h-5" />
              </div>
              <div>
                <h4 class="font-bold text-gray-900 dark:text-white text-base">{{ selectedConnForSchema.name }}</h4>
                <p class="text-xs text-gray-500 dark:text-gray-400 font-mono">
                  {{ selectedConnForSchema.host }}:{{ selectedConnForSchema.port }} / {{ selectedConnForSchema.database }} ({{ selectedConnForSchema.environment }})
                </p>
              </div>
            </div>

            <!-- Stats counter -->
            <div class="flex items-center gap-4 text-xs">
              <div class="text-center px-3 py-1.5 rounded-lg bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
                <span class="block font-bold text-gray-900 dark:text-white text-sm">{{ discoveredTables.length }}</span>
                <span class="text-gray-500 dark:text-gray-400">{{ t('settings.dataSource.tableCount') }}</span>
              </div>
              <div class="text-center px-3 py-1.5 rounded-lg bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
                <span class="block font-bold text-emerald-600 dark:text-emerald-400 text-sm">{{ activeMappingsCount }}</span>
                <span class="text-gray-500 dark:text-gray-400">{{ t('settings.dataSource.mappedCount') }}</span>
              </div>
            </div>
          </div>

          <!-- Filter & Search Tables -->
          <div class="flex flex-col sm:flex-row items-center justify-between gap-3">
            <div class="w-full sm:w-80">
              <CommonSearchInput
                v-model="tableSearchQuery"
                :placeholder="t('settings.dataSource.searchTablesPlaceholder')"
                size="md"
                class="w-full"
              />
            </div>
            <div class="flex items-center gap-2 w-full sm:w-auto justify-end">
              <USelect
                v-model="scopeFilter"
                :items="['All Scopes', 'Risk Management', 'Audit Features', 'QAR Features', 'Data Analytics / AI']"
                size="md"
                class="w-48"
              />
              <UButton
                color="warning"
                variant="subtle"
                size="md"
                icon="i-lucide-shield-alert"
                :loading="isReconcilingDeletes"
                class="rounded-xl font-semibold text-xs shrink-0"
                @click="triggerDeleteReconciliation"
              > {{ t('settings.dataSource.mapping.deletionDetection') }} </UButton>
            </div>
          </div>

          <!-- Loading State -->
          <div v-if="isLoadingSchema" class="text-center py-12 space-y-3">
            <UIcon name="i-lucide-loader-2" class="w-8 h-8 animate-spin text-primary-500 mx-auto" />
            <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('settings.dataSource.mapping.introspecting') }}</p>
          </div>

          <!-- Discovered Tables & Mappings Accordion / List -->
          <div v-else class="space-y-4">
            <div
              v-for="table in filteredTables"
              :key="table.tableName"
              class="border border-gray-200 dark:border-gray-800 rounded-xl p-4 bg-white dark:bg-gray-900 transition-all hover:border-primary-500/40 shadow-sm space-y-3"
            >
              <div class="flex flex-col md:flex-row md:items-center justify-between gap-3">
                <!-- Table metadata -->
                <div class="space-y-1">
                  <div class="flex items-center gap-2.5 flex-wrap">
                    <UIcon name="i-lucide-table" class="w-4 h-4 text-blue-500" />
                    <span class="font-mono font-bold text-gray-900 dark:text-white text-base">{{ table.tableName }}</span>
                    <UBadge
                      :color="getTableArchetypeInfo(table).meta.color"
                      variant="subtle"
                      size="sm"
                      class="rounded-md font-semibold text-[11px]"
                    >
                      {{ getTableArchetypeInfo(table).meta.label }}
                    </UBadge>
                    <UBadge color="neutral" variant="subtle" size="sm" class="rounded-md font-mono text-[11px]">
                      {{ table.rowCount.toLocaleString() }} {{ t('settings.dataSource.mapping.rows') }}
                    </UBadge>
                    <UBadge color="neutral" variant="outline" size="sm" class="rounded-md font-mono text-[11px]">
                      {{ table.columns.length }} {{ t('settings.dataSource.mapping.columnsShort') }}
                    </UBadge>
                    <UBadge
                      v-if="getTableMappings(table.tableName).length > 0"
                      color="success"
                      variant="subtle"
                      size="sm"
                      class="rounded-md font-semibold text-[11px]"
                    >
                      {{ getTableMappings(table.tableName).length }} {{ t('settings.dataSource.mappedCount') }}
                    </UBadge>
                  </div>
                  <p v-if="table.description" class="text-xs text-gray-500 dark:text-gray-400">
                    {{ table.description }}
                  </p>
                  <p
                    v-if="getTableRecommendedModules(table).length > 0"
                    class="text-[11px] text-primary-700 dark:text-primary-400 flex items-center gap-1.5 flex-wrap"
                  >
                    <UIcon name="i-lucide-sparkles" class="w-3.5 h-3.5 shrink-0" />
                    <span class="font-semibold">Direkomendasikan:</span>
                    <span>{{ getTableRecommendedModules(table).map(r => r.targetModule).join(', ') }}</span>
                  </p>
                </div>

                <!-- Table Quick Actions -->
                <div class="flex items-center gap-2 shrink-0 flex-wrap">
                  <UButton
                    color="neutral"
                    variant="subtle"
                    size="sm"
                    icon="i-lucide-eye"
                    class="rounded-lg font-medium text-xs"
                    @click="openPreviewModal(table.tableName)"
                  >
                    {{ t('settings.dataSource.previewData') }}
                  </UButton>
                  <UButton
                    v-if="getTableRecommendedModules(table).length > 0"
                    color="success"
                    variant="subtle"
                    size="sm"
                    icon="i-lucide-sparkles"
                    class="rounded-lg font-semibold text-xs"
                    :disabled="getUnmappedRecommendedCount(table) === 0"
                    @click="applyRecommendedMappings(table)"
                  >
                    {{ getUnmappedRecommendedCount(table) > 0 ? `Terapkan Rekomendasi (${getUnmappedRecommendedCount(table)})` : 'Semua Rekomendasi Terpasang' }}
                  </UButton>
                  <UButton
                    color="primary"
                    variant="subtle"
                    size="sm"
                    icon="i-lucide-plus"
                    class="rounded-lg font-semibold text-xs"
                    @click="addTableMapping(table)"
                  >
                    + Add Mapping
                  </UButton>
                </div>
              </div>

              <!-- Columns Chips preview -->
              <div class="pt-2 border-t border-gray-100 dark:border-gray-800 flex items-center gap-1.5 flex-wrap">
                <span class="text-[11px] font-semibold text-gray-400 mr-1">{{ t('settings.dataSource.mapping.columns') }}</span>
                <span
                  v-for="col in table.columns"
                  :key="col.name"
                  class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-mono bg-gray-50 dark:bg-gray-800/80 border border-gray-200 dark:border-gray-700/60"
                  :class="col.isPrimary ? 'text-amber-600 dark:text-amber-400 font-bold border-amber-300 dark:border-amber-900/50' : 'text-gray-700 dark:text-gray-300'"
                >
                  <UIcon v-if="col.isPrimary" name="i-lucide-key" class="w-3 h-3 text-amber-500" />
                  {{ col.name }} <span class="text-[9px] text-gray-400">({{ col.dataType }})</span>
                </span>
              </div>

              <!-- Mapping Configuration Form Cards (Multi-mapping per table) -->
              <div
                v-for="(mapping, mIdx) in getTableMappings(table.tableName)"
                :key="mIdx"
                class="mt-3 p-3.5 rounded-xl border space-y-3"
                :class="getMappingCompatibility(table, mapping.targetModule).compatible
                  ? 'bg-primary-50/50 dark:bg-primary-950/20 border-primary-200/60 dark:border-primary-900/40'
                  : 'bg-amber-50/60 dark:bg-amber-950/20 border-amber-300/80 dark:border-amber-800/60'"
              >
                <div class="flex items-center justify-between gap-2 flex-wrap">
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class="text-xs font-bold text-primary-900 dark:text-primary-300 flex items-center gap-1.5">
                      <UIcon name="i-lucide-settings-2" class="w-3.5 h-3.5 text-primary-500" />
                      Mapping #{{ mIdx + 1 }} &mdash; {{ mapping.targetModule || getScopeLabel(mapping.targetScope) }}
                    </span>
                    <UBadge
                      v-if="getMappingCompatibility(table, mapping.targetModule).compatible"
                      :color="getImportModeBadge(getMappingCompatibility(table, mapping.targetModule).importMode).color"
                      variant="subtle"
                      size="xs"
                      class="font-semibold text-[10px]"
                    >
                      {{ getImportModeBadge(getMappingCompatibility(table, mapping.targetModule).importMode).label }}
                    </UBadge>
                    <UBadge
                      v-if="selectedConnForSchema && getMappingImportStatus(selectedConnForSchema.id, table.tableName, mapping.targetModule)"
                      :color="getMappingImportStatus(selectedConnForSchema.id, table.tableName, mapping.targetModule)?.status === 'imported'
                        ? 'success'
                        : getMappingImportStatus(selectedConnForSchema.id, table.tableName, mapping.targetModule)?.status === 'manual_required'
                          ? 'warning'
                          : getMappingImportStatus(selectedConnForSchema.id, table.tableName, mapping.targetModule)?.status === 'skipped_duplicate'
                            ? 'neutral'
                            : 'error'"
                      variant="outline"
                      size="xs"
                      class="font-mono text-[10px]"
                    >
                      <template v-if="getMappingImportStatus(selectedConnForSchema.id, table.tableName, mapping.targetModule)?.status === 'imported'">
                        ✓ +{{ getMappingImportStatus(selectedConnForSchema.id, table.tableName, mapping.targetModule)?.addedCount }} diimpor ({{ formatImportTime(getMappingImportStatus(selectedConnForSchema.id, table.tableName, mapping.targetModule)?.importedAt) }})
                      </template>
                      <template v-else-if="getMappingImportStatus(selectedConnForSchema.id, table.tableName, mapping.targetModule)?.status === 'skipped_duplicate'">
                        • Sudah sinkron ({{ formatImportTime(getMappingImportStatus(selectedConnForSchema.id, table.tableName, mapping.targetModule)?.importedAt) }})
                      </template>
                      <template v-else-if="getMappingImportStatus(selectedConnForSchema.id, table.tableName, mapping.targetModule)?.status === 'manual_required'">
                        ⚑ Konfirmasi via Banner
                      </template>
                      <template v-else>
                        ! {{ getMappingImportStatus(selectedConnForSchema.id, table.tableName, mapping.targetModule)?.status }}
                      </template>
                    </UBadge>
                  </div>
                  <div class="flex items-center gap-2">
                    <span class="text-[11px] font-mono text-primary-700 dark:text-primary-400 hidden sm:inline">{{ t('settings.dataSource.mapping.targetSchema') }} {{ selectedConnForSchema.database }}</span>
                    <UButton
                      color="error"
                      variant="ghost"
                      size="xs"
                      icon="i-lucide-trash-2"
                      class="rounded-lg text-[11px] font-medium"
                      @click="removeTableMapping(table.tableName, mIdx)"
                    > Remove </UButton>
                  </div>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div>
                    <label class="block text-[11px] font-bold text-gray-700 dark:text-gray-300 mb-1">
                      {{ t('settings.dataSource.targetScope') }}
                    </label>
                    <USelect
                      :model-value="mapping.targetScope"
                      :items="getScopesForTable(table)"
                      size="sm"
                      class="w-full"
                      @update:model-value="(val) => updateMappingField(table.tableName, mIdx, 'targetScope', val)"
                    />
                  </div>

                  <div>
                    <label class="block text-[11px] font-bold text-gray-700 dark:text-gray-300 mb-1">
                      {{ t('settings.dataSource.targetModule') }}
                    </label>
                    <USelect
                      :model-value="mapping.targetModule"
                      :items="getModulesForScope(mapping.targetScope, table)"
                      size="sm"
                      class="w-full text-xs font-medium"
                      @update:model-value="(val) => updateMappingField(table.tableName, mIdx, 'targetModule', val)"
                    />
                  </div>
                </div>

                <!-- Compatibility Status & Reason Note -->
                <div
                  v-if="getMappingCompatibility(table, mapping.targetModule).compatible"
                  class="flex items-start gap-2 text-[11px] text-gray-600 dark:text-gray-300 bg-white/70 dark:bg-gray-900/50 border border-gray-200/70 dark:border-gray-800 rounded-lg px-2.5 py-1.5"
                >
                  <UIcon name="i-lucide-check-circle-2" class="w-3.5 h-3.5 text-emerald-500 shrink-0 mt-0.5" />
                  <span>{{ getMappingCompatibility(table, mapping.targetModule).reason }}</span>
                </div>
                <div
                  v-else
                  class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-[11px] text-amber-800 dark:text-amber-300 bg-amber-100/70 dark:bg-amber-950/50 border border-amber-300 dark:border-amber-800 rounded-lg px-3 py-2"
                >
                  <div class="flex items-start gap-2">
                    <UIcon name="i-lucide-alert-triangle" class="w-4 h-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5" />
                    <span>{{ getMappingCompatibility(table, mapping.targetModule).reason }}</span>
                  </div>
                  <UButton
                    size="xs"
                    color="warning"
                    variant="solid"
                    icon="i-lucide-sparkles"
                    class="rounded-lg font-semibold shrink-0"
                    @click="fixIncompatibleMapping(table, mIdx)"
                  >
                    Ganti ke Rekomendasi
                  </UButton>
                </div>

                <!-- Anomaly Detection Rules Checklist & AI Training Section - only displayed when module is AI: Anomaly Detection -->
                <div v-if="shouldShowAnomalyRules(mapping.targetScope, mapping.targetModule)" class="pt-3 border-t border-primary-200/50 dark:border-primary-900/50 space-y-3">
                  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                    <div>
                      <div class="flex items-center gap-2">
                        <label class="block text-xs font-bold text-gray-800 dark:text-gray-200">
                          {{ t('settings.dataSource.anomalyRules') }}
                        </label>
                        <UBadge color="primary" variant="subtle" size="xs" class="font-mono text-[10px]">
                          AI Model Trainer
                        </UBadge>
                      </div>
                      <p class="text-[11px] text-gray-500 dark:text-gray-400 mt-0.5">
                        Pilih aturan preset atau buat aturan kustom untuk melatih model AI Isolation Forest & XGBoost secara otomatis.
                      </p>
                    </div>

                    <!-- Action: Latih Model AI Button -->
                    <div class="flex items-center gap-2 flex-wrap shrink-0 self-start sm:self-auto">
                      <UButton
                        size="xs"
                        color="primary"
                        variant="solid"
                        icon="i-lucide-save"
                        :loading="isTrainingModel[getMappingKey(table.tableName, mIdx)]"
                        class="font-semibold shadow-xs rounded-lg"
                        @click="trainModelForMapping(table, mIdx)"
                      >
                        Terapkan Aturan Anomaly
                      </UButton>
                    </div>
                  </div>

                  <!-- Standard Preset Rules Chips -->
                  <div class="space-y-1.5">
                    <div class="text-[10px] font-bold uppercase tracking-wider text-gray-400">
                      Aturan Anomali Standar (Preset)
                    </div>
                    <div class="flex items-center gap-1.5 flex-wrap">
                      <button
                        v-for="rule in standardAnomalyRules"
                        :key="rule"
                        type="button"
                        class="px-2.5 py-1 rounded-lg text-[11px] font-medium transition-all flex items-center gap-1.5"
                        :class="mapping.anomalyRules?.includes(rule)
                          ? 'bg-primary-600 text-white font-semibold shadow-xs'
                          : 'bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:border-gray-300'"
                        @click="toggleAnomalyRule(table.tableName, mIdx, rule)"
                      >
                        <UIcon
                          :name="mapping.anomalyRules?.includes(rule) ? 'i-lucide-check' : 'i-lucide-plus'"
                          class="w-3 h-3"
                        />
                        {{ rule }}
                      </button>
                    </div>
                  </div>

                  <!-- Custom Rules Section -->
                  <div class="space-y-2 pt-1">
                    <div class="flex items-center justify-between">
                      <div class="text-[10px] font-bold uppercase tracking-wider text-gray-400">
                        Aturan Anomali Kustom ({{ (mapping.customRules || []).length }})
                      </div>
                      <UButton
                        v-if="activeRuleBuilderKey !== getMappingKey(table.tableName, mIdx)"
                        size="xs"
                        variant="outline"
                        color="neutral"
                        icon="i-lucide-plus"
                        class="text-[11px] py-0.5 rounded-lg"
                        @click="openCustomRuleBuilder(table.tableName, mIdx, table)"
                      >
                        Tambah Aturan Kustom
                      </UButton>
                    </div>

                    <!-- Active Custom Rules Badges/List -->
                    <div v-if="mapping.customRules && mapping.customRules.length > 0" class="space-y-1.5">
                      <div
                        v-for="cRule in mapping.customRules"
                        :key="cRule.id"
                        class="flex items-center justify-between px-2.5 py-1.5 rounded-lg border text-xs bg-gray-50/70 dark:bg-gray-800/60 border-gray-200 dark:border-gray-700/80"
                      >
                        <div class="flex items-center gap-2 overflow-hidden">
                          <span
                            :class="['inline-flex items-center px-2 py-0.5 rounded font-bold text-[10px] shrink-0', getRiskLevelColorClass(normalizeAnomalyRiskLevel(cRule.severity))]"
                          >
                            {{ normalizeAnomalyRiskLevel(cRule.severity) }}
                          </span>
                          <span class="font-semibold text-gray-900 dark:text-white truncate">{{ cRule.name }}</span>
                          <span class="text-gray-400 text-[11px] font-mono">({{ cRule.field }} {{ cRule.operator }} {{ cRule.thresholdStr || cRule.threshold }})</span>
                        </div>
                        <div class="flex items-center gap-1.5 shrink-0">
                          <button
                            type="button"
                            class="p-1 text-gray-400 hover:text-rose-500 transition-colors"
                            title="Hapus Aturan"
                            @click="removeCustomRule(table.tableName, mIdx, cRule.id)"
                          >
                            <UIcon name="i-lucide-trash-2" class="w-3.5 h-3.5" />
                          </button>
                        </div>
                      </div>
                    </div>
                    <div v-else-if="activeRuleBuilderKey !== getMappingKey(table.tableName, mIdx)" class="text-[11px] text-gray-400 italic">
                      Belum ada aturan kustom. Klik "+ Tambah Aturan Kustom" untuk menambahkan kriteria khusus seperti ambang batas nominal, pola transaksi, atau status tertentu.
                    </div>

                    <!-- Inline Custom Rule Builder Drawer / Form -->
                    <div
                      v-if="activeRuleBuilderKey === getMappingKey(table.tableName, mIdx)"
                      class="p-3 rounded-xl border border-primary-200 dark:border-primary-800 bg-primary-50/30 dark:bg-primary-950/20 space-y-2.5"
                    >
                      <div class="flex items-center justify-between pb-1 border-b border-primary-100 dark:border-primary-900/50">
                        <span class="text-xs font-bold text-primary-900 dark:text-primary-100 flex items-center gap-1.5">
                          <UIcon name="i-lucide-sliders" class="w-3.5 h-3.5" />
                          Konfigurasi Aturan Anomali Baru
                        </span>
                        <button
                          type="button"
                          class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                          @click="closeCustomRuleBuilder"
                        >
                          <UIcon name="i-lucide-x" class="w-3.5 h-3.5" />
                        </button>
                      </div>

                      <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                        <div>
                          <label class="block text-[10px] font-bold text-gray-600 dark:text-gray-300 mb-0.5">Nama Aturan Anomali</label>
                          <UInput
                            v-model="newCustomRule.name"
                            placeholder="misal: Transaksi Luar Batas Otorisasi"
                            size="xs"
                            class="w-full"
                          />
                        </div>
                        <div>
                          <label class="block text-[10px] font-bold text-gray-600 dark:text-gray-300 mb-0.5">Tingkat Keparahan (Severity)</label>
                          <USelect
                            v-model="newCustomRule.severity"
                            :items="ruleSeverityOptions"
                            size="xs"
                            class="w-full font-mono text-xs"
                          />
                        </div>
                      </div>

                      <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
                        <div>
                          <label class="block text-[10px] font-bold text-gray-600 dark:text-gray-300 mb-0.5">Kolom Target</label>
                          <USelect
                            v-model="newCustomRule.field"
                            :items="getTableColumnOptions(table)"
                            size="xs"
                            class="w-full font-mono text-xs"
                          />
                        </div>
                        <div>
                          <label class="block text-[10px] font-bold text-gray-600 dark:text-gray-300 mb-0.5">Operator Kondisi</label>
                          <USelect
                            v-model="newCustomRule.operator"
                            :items="ruleOperatorOptions"
                            size="xs"
                            class="w-full font-mono text-xs"
                          />
                        </div>
                        <div>
                          <label class="block text-[10px] font-bold text-gray-600 dark:text-gray-300 mb-0.5">Nilai Ambang (Threshold)</label>
                          <UInput
                            v-model="newCustomRule.threshold"
                            type="number"
                            placeholder="100000000"
                            size="xs"
                            class="w-full font-mono text-xs"
                          />
                        </div>
                      </div>

                      <div>
                        <label class="block text-[10px] font-bold text-gray-600 dark:text-gray-300 mb-0.5">Deskripsi / Indikasi Pelanggaran Audit</label>
                        <UInput
                          v-model="newCustomRule.description"
                          placeholder="Alasan anomali, misal: Transaksi melampaui batas otorisasi cabang tanpa persetujuan komite"
                          size="xs"
                          class="w-full"
                        />
                      </div>

                      <div class="flex items-center justify-end gap-2 pt-1">
                        <UButton size="xs" variant="ghost" color="neutral" class="rounded-lg" @click="closeCustomRuleBuilder">
                          Batal
                        </UButton>
                        <UButton size="xs" color="primary" icon="i-lucide-check" class="rounded-lg font-semibold" @click="saveCustomRule(table.tableName, mIdx)">
                          Simpan Aturan
                        </UButton>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="filteredTables.length === 0" class="text-center py-8 text-gray-500 dark:text-gray-400 text-sm">
              {{ t('settings.dataSource.noTablesFound') }}
            </div>
          </div>

          <!-- Bulk Import Summary Panel inside Schema Modal -->
          <div
            v-if="showImportSummaryPanel && lastBulkImportSummary && lastBulkImportSummary.connectionId === selectedConnForSchema.id"
            class="p-4 rounded-xl border border-emerald-200 dark:border-emerald-900/60 bg-emerald-50/40 dark:bg-emerald-950/20 space-y-3"
          >
            <div class="flex items-center justify-between gap-2">
              <div class="flex items-center gap-2 flex-wrap">
                <UIcon name="i-lucide-check-circle-2" class="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
                <span class="text-xs font-bold text-gray-900 dark:text-white">
                  Ringkasan Hasil Impor ({{ lastBulkImportSummary.totalMappings }} mapping diproses)
                </span>
                <UBadge color="success" variant="subtle" size="xs" class="font-mono">
                  +{{ lastBulkImportSummary.totalRecordsAdded }} data baru
                </UBadge>
                <UBadge v-if="lastBulkImportSummary.manualMappings > 0" color="warning" variant="subtle" size="xs">
                  {{ lastBulkImportSummary.manualMappings }} konfirmasi manual
                </UBadge>
              </div>
              <UButton
                color="neutral"
                variant="ghost"
                size="xs"
                icon="i-lucide-x"
                @click="showImportSummaryPanel = false"
              />
            </div>

            <div class="space-y-1.5 max-h-56 overflow-y-auto pr-1">
              <div
                v-for="res in lastBulkImportSummary.results"
                :key="res.key"
                class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 px-3 py-2 rounded-lg bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 text-xs"
              >
                <div class="min-w-0 space-y-0.5">
                  <div class="flex items-center gap-1.5 flex-wrap">
                    <span class="font-mono font-bold text-gray-900 dark:text-white">{{ res.tableName }}</span>
                    <span class="text-gray-400">→</span>
                    <span class="font-semibold text-gray-800 dark:text-gray-200">{{ res.targetModule }}</span>
                    <UBadge
                      :color="res.status === 'imported' ? 'success' : res.status === 'manual_required' ? 'warning' : res.status === 'skipped_duplicate' ? 'neutral' : 'error'"
                      variant="subtle"
                      size="xs"
                    >
                      {{ res.status === 'imported' ? `+${res.addedCount} Diimpor` : res.status === 'manual_required' ? 'Lewat Banner' : res.status === 'skipped_duplicate' ? 'Sudah Sinkron' : res.status === 'incompatible' ? 'Tidak Cocok' : 'Gagal' }}
                    </UBadge>
                  </div>
                  <p class="text-[11px] text-gray-600 dark:text-gray-400">{{ res.message }}</p>
                </div>
                <UButton
                  v-if="res.route && res.route !== '/' && res.status !== 'incompatible'"
                  :to="res.route"
                  color="primary"
                  variant="subtle"
                  size="xs"
                  icon="i-lucide-external-link"
                  class="rounded-lg font-semibold shrink-0"
                  @click="closeSchemaModal()"
                >
                  {{ res.status === 'manual_required' ? 'Buka Banner' : 'Buka' }}
                </UButton>
              </div>
            </div>
          </div>

          <!-- Footer Save & Bulk Import Buttons -->
          <div class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-2 pt-4 border-t border-gray-200 dark:border-gray-800">
            <UButton type="button" color="neutral" variant="subtle" class="rounded-xl font-semibold justify-center" @click="closeSchemaModal">
              {{ t('settings.dataSource.close') }}
            </UButton>

            <div class="flex items-center gap-2 flex-wrap justify-end">
              <UButton
                type="button"
                color="success"
                variant="subtle"
                icon="i-lucide-download"
                class="rounded-xl font-bold"
                :loading="isImportingMappings"
                :disabled="activeMappingsCount === 0 || isSavingMappings || isImportingMappings"
                @click="saveAndImportAllMappings"
              >
                Impor Semua Mapping ({{ activeMappingsCount }})
              </UButton>

              <UButton
                type="button"
                color="primary"
                icon="i-lucide-save"
                class="rounded-xl font-bold"
                :loading="isSavingMappings"
                :disabled="isImportingMappings"
                @click="saveSchemaMappings"
              >
                {{ t('settings.dataSource.saveMappings') }}
              </UButton>
            </div>
          </div>
        </div>
      </template>
    </UModal>

    <!-- Data Preview Modal -->
    <UModal
      v-model:open="isPreviewModalOpen"
      :title="`${t('settings.dataSource.sampleRecordsTitle')}: ${previewTableName}`"
      :description="`${t('settings.dataSource.sampleRecordsDesc')} ${previewTableName} (${selectedConnForSchema?.name || ''})`"
      :close="{ color: 'neutral', variant: 'outline', onClick: () => { isPreviewModalOpen = false } }"
      dismissible
      class="w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl"
      :ui="{
        content: 'bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-2xl overflow-hidden',
        header: 'px-6 py-5 border-b border-gray-100 dark:border-gray-800 pr-14',
        title: 'text-lg font-bold text-gray-900 dark:text-white leading-snug',
        description: 'text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-1.5 leading-relaxed',
        body: 'p-4 sm:p-6 space-y-4 bg-white dark:bg-gray-900 max-h-[75vh] overflow-y-auto',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
    >
      <template #body>
        <div v-if="isLoadingPreview" class="text-center py-12 space-y-3">
          <UIcon name="i-lucide-loader-2" class="w-8 h-8 animate-spin text-primary-500 mx-auto" />
          <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('settings.dataSource.mapping.fetchingRows') }}</p>
        </div>
        <div v-else-if="previewRows.length > 0 && previewRows[0]" class="overflow-x-auto border border-gray-200 dark:border-gray-800 rounded-xl">
          <table class="w-full text-xs text-left">
            <thead class="bg-gray-50 dark:bg-gray-800/80 text-gray-600 dark:text-gray-400 font-semibold border-b border-gray-200 dark:border-gray-800">
              <tr>
                <th v-for="col in Object.keys(previewRows[0] || {})" :key="col" class="py-2.5 px-3 font-mono">
                  {{ col }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800/60 font-mono text-gray-700 dark:text-gray-300">
              <tr v-for="(row, rIdx) in previewRows" :key="rIdx" class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
                <td v-for="(val, k) in row" :key="k" class="py-2 px-3 whitespace-nowrap">
                  <span v-if="val === null" class="text-gray-400 italic">null</span>
                  <span v-else>{{ val }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="text-center py-8 text-gray-500 dark:text-gray-400 text-sm">
          {{ t('settings.dataSource.noDataPreview') }}
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { getMasterServiceBaseUrl, getDataHubBaseUrl } from '~/composables/useApiUrl'
import { useDataSourceStore } from '~/stores/data-source'
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import { getCanonicalCategory } from '~/composables/useAnomalyScatter'
import { normalizeAnomalyRiskLevel, toAnomalyRiskCode, getRiskLevelColorClass } from '~/utils/riskLevelBadge'
import {
  SYNC_FREQUENCY_OPTIONS,
  type SyncFrequencyOption,
  analyzeTableSchema,
  evaluateModuleCompatibility,
  getDefaultModuleForTargetScope,
  getModulesForTargetScope,
  getNextRecommendedMappingForTable,
  getRecommendedMappingsForTable,
  getScopeOptionsForTable,
  getTableArchetypeMeta,
  normalizeSyncSchedule,
  resolveAutoMappingMetadata,
  shouldShowAnomalyRules,
  type ModuleCompatibilityEvaluation,
} from '~/utils/dataSourceConfig'
import {
  useMappedDataImporter,
  type BulkImportSummary,
} from '~/composables/useMappedDataImporter'
import SourceSelectorDropdown from '~/components/common/SourceSelectorDropdown.vue'

const toast = useToast()
const { t } = useI18n()
const config = useRuntimeConfig()
const dataSourceStore = useDataSourceStore()
const { trainAnomalyWithRules } = useAiAnalytics()
const {
  isImporting: isImportingMappings,
  lastSummary: lastBulkImportSummary,
  getMappingStatus: getMappingImportStatus,
  importAllForConnection,
} = useMappedDataImporter()
const importingConnId = ref<string | null>(null)
const showImportSummaryPanel = ref(false)

export interface TableColumn {
  name: string
  dataType: string
  isNullable: boolean
  isPrimary: boolean
}

export interface SchemaTable {
  tableName: string
  rowCount: number
  columnCount: number
  columns: TableColumn[]
  description?: string
}

export interface CustomAnomalyRule {
  id: string
  name: string
  field: string
  operator: string
  threshold: number
  thresholdStr?: string
  severity: 'HIGH' | 'MODERATE_HIGH' | 'MODERATE' | 'LOW_MODERATE' | 'LOW'
  category?: string
  description?: string
  isActive?: boolean
  trainedAt?: string
}

export interface TableDataMapping {
  tableName: string
  targetScope: string
  targetModule: string
  auditField: string
  canonicalEntity?: string
  watermarkColumn?: string
  softDeleteColumn?: string
  enableDeleteDetection?: boolean
  columnMappings?: Record<string, string>
  syncInterval: string
  anomalyRules?: string[]
  customRules?: CustomAnomalyRule[]
  isActive: boolean
}

export interface DataSourceConn {
  id: string
  name: string
  type: 'postgres' | 'mysql' | 'oracle' | 'sap' | 'rest' | 's3' | 'mssql' | 'kafka'
  host: string
  port: number
  database: string
  environment: 'Production' | 'Staging' | 'Development'
  status: 'Connected' | 'Syncing' | 'Error' | 'Inactive'
  syncSchedule: string
  lastSync: string
  lastSyncedAt?: string
  nextSyncAt?: string
  lastError?: string
  ssl: boolean
  username?: string
  password?: string
  scopes?: string[]
  dataMappings?: TableDataMapping[]
}

const availableScopes = [
  { id: 'risk_management', label: '1. Risk Management', description: 'Risk Profile, Heatmap, RCM & Appetite' },
  { id: 'audit_features', label: '2. Audit Features', description: 'Activity Plan, Working Paper, ATR & Reports' },
  { id: 'qar_features', label: '3. QAR Features', description: 'Quality Assurance Review & Compliance' },
  { id: 'data_analytics', label: '4. Data Analytics / AI', description: 'AI Training, Anomaly Detection, Risk Scoring & Predictive Analytics' },
]

const standardAnomalyRules = [
  'Threshold Breach (>100M)',
  'Weekend / After-hours Activity',
  'Duplicate Invoices / Payments',
  'Privilege Escalation Alert',
  'Rapid Turnover in Dormant Account',
  'High Frequency Velocity (Rapid Bursts)',
  'Split Transactions Below Limit (Smurfing)',
  'Sudden Balance Depletion (>80% Outflow)'
]

export interface ActivityLog {
  id: string
  connName: string
  type: DataSourceConn['type']
  event: string
  status: string
  records: number
  duration: string
  timestamp: string
}

// Data Sources (empty by default)
const connections = ref<DataSourceConn[]>([])

// Activity Logs (empty by default)
const activityLogs = ref<ActivityLog[]>([])

function formatSyncDateTime(isoStr?: string): string {
  if (!isoStr) return ''
  const d = new Date(isoStr)
  if (Number.isNaN(d.getTime())) return isoStr
  return d.toLocaleString('id-ID', {
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatConnLastSync(conn: DataSourceConn): string {
  if (conn.status === 'Syncing') return conn.lastSync
  if (conn.lastSyncedAt) return formatSyncDateTime(conn.lastSyncedAt)
  return conn.lastSync
}

function normalizeConnStatus(status?: string): DataSourceConn['status'] {
  const s = (status || '').trim().toLowerCase()
  if (s === 'connected' || s === 'synced' || s === 'registered') return 'Connected'
  if (s === 'syncing' || s === 'running' || s === 'queued') return 'Syncing'
  if (s === 'error' || s === 'failed') return 'Error'
  if (s === 'inactive' || s === 'disconnected') return 'Inactive'
  return 'Connected'
}

// Load connections & logs from backend API on mount
async function fetchConnectionsFromApi() {
  try {
    const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources`)
    if (res && res.data && Array.isArray(res.data) && res.data.length > 0) {
      const normalized = res.data.map((item: any) => ({
        ...item,
        status: normalizeConnStatus(item.status),
        syncSchedule: normalizeSyncSchedule(item.syncSchedule),
      }))
      connections.value = normalized
      dataSourceStore.setSources(normalized)
    } else if (dataSourceStore.sources.length > 0) {
      connections.value = dataSourceStore.sources.map((item: any) => ({
        ...item,
        status: normalizeConnStatus(item.status),
        syncSchedule: normalizeSyncSchedule(item.syncSchedule),
      })) as any
    }
  } catch (err) {
    if (dataSourceStore.sources.length > 0) {
      connections.value = dataSourceStore.sources.map((item: any) => ({
        ...item,
        status: normalizeConnStatus(item.status),
        syncSchedule: normalizeSyncSchedule(item.syncSchedule),
      })) as any
    }
    console.warn('Backend API connection unavailable, operating in local mode:', err)
  }
}

async function fetchLogsFromApi() {
  try {
    const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources/logs`)
    if (res && res.data) {
      activityLogs.value = res.data
    }
  } catch (err) {
    console.warn('Backend API logs unavailable:', err)
  }
}

const SYNC_COOLDOWN_SEC = 60
const COOLDOWN_STORAGE_PREFIX = 'ds_sync_cooldown_'
const cooldownNow = ref(Date.now())
const syncCooldownEnds = ref<Record<string, number>>({})
let cooldownTimer: any = null
let scheduledRefreshTimer: any = null

function loadCooldownsFromStorage() {
  if (typeof window === 'undefined') return
  const loaded: Record<string, number> = {}
  const now = Date.now()
  for (let i = 0; i < window.localStorage.length; i++) {
    const k = window.localStorage.key(i)
    if (k && k.startsWith(COOLDOWN_STORAGE_PREFIX)) {
      const targetKey = k.slice(COOLDOWN_STORAGE_PREFIX.length)
      const expiresAt = Number(window.localStorage.getItem(k) || '0')
      if (expiresAt > now) {
        loaded[targetKey] = expiresAt
      } else {
        window.localStorage.removeItem(k)
      }
    }
  }
  syncCooldownEnds.value = loaded
}

function getSyncCooldownRemaining(key: string): number {
  const expiresAt = syncCooldownEnds.value[key] || 0
  if (!expiresAt) return 0
  const rem = Math.ceil((expiresAt - cooldownNow.value) / 1000)
  return rem > 0 ? rem : 0
}

function setSyncCooldown(key: string, seconds: number = SYNC_COOLDOWN_SEC) {
  const sec = Math.max(1, Number(seconds) || SYNC_COOLDOWN_SEC)
  const expiresAt = Date.now() + sec * 1000
  syncCooldownEnds.value = { ...syncCooldownEnds.value, [key]: expiresAt }
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(`${COOLDOWN_STORAGE_PREFIX}${key}`, String(expiresAt))
  }
}

onMounted(() => {
  loadCooldownsFromStorage()
  cooldownTimer = setInterval(() => {
    cooldownNow.value = Date.now()
  }, 1000)
  scheduledRefreshTimer = setInterval(() => {
    if (!isBatchSyncing.value && !syncingId.value) {
      fetchConnectionsFromApi()
      fetchLogsFromApi()
    }
  }, 30000)
  fetchConnectionsFromApi()
  fetchLogsFromApi()
})

onBeforeUnmount(() => {
  if (cooldownTimer) clearInterval(cooldownTimer)
  if (scheduledRefreshTimer) clearInterval(scheduledRefreshTimer)
  if (batchPollInterval) clearInterval(batchPollInterval)
})

// Filter state
const searchQuery = ref('')
const statusFilter = ref('ALL')
const typeFilter = ref('ALL')

// Synchronized dynamically with actual connection statuses
const statusOptions = computed(() => {
  const validStatuses = new Set<string>(['Connected', 'Syncing', 'Error'])
  for (const c of connections.value) {
    if (c.status) validStatuses.add(normalizeConnStatus(c.status))
  }
  return [
    { label: t('settings.dataSource.allStatuses'), value: 'ALL' },
    ...Array.from(validStatuses).map(st => ({ label: st, value: st })),
  ]
})

// Synchronized dynamically with available providers (currently PostgreSQL, auto-expands as providers are added)
const typeOptions = computed(() => {
  const availableProviders = new Set<string>(providers.map(p => p.name))
  for (const c of connections.value) {
    if (c.type) availableProviders.add(getProviderName(c.type))
  }
  return [
    { label: t('settings.dataSource.allProviders'), value: 'ALL' },
    ...Array.from(availableProviders).map(name => ({ label: name, value: name })),
  ]
})

// Filtered list
const filteredConnections = computed(() => {
  return connections.value.filter(c => {
    const q = searchQuery.value.toLowerCase()
    const matchesSearch = (c.name || '').toLowerCase().includes(q) ||
                          (c.host || '').toLowerCase().includes(q) ||
                          (c.database || '').toLowerCase().includes(q)

    const connStatus = normalizeConnStatus(c.status)
    const matchesStatus = statusFilter.value === 'ALL' || connStatus === statusFilter.value
    const matchesType = typeFilter.value === 'ALL' || getProviderName(c.type) === typeFilter.value

    return matchesSearch && matchesStatus && matchesType
  })
})

function resetFilters() {
  searchQuery.value = ''
  statusFilter.value = 'ALL'
  typeFilter.value = 'ALL'
}

// Providers metadata (PostgreSQL Only)
const providers = [
  { id: 'postgres', name: 'PostgreSQL', icon: 'i-lucide-database', defaultPort: 5432, badge: 'Standard Connector', color: 'text-blue-600 bg-blue-50 border-blue-200 dark:bg-blue-950/40 dark:border-blue-900/50' },
]

function getProviderIcon(type: string) {
  const found = providers.find(p => p.id === type)
  return found ? found.icon : 'i-lucide-database'
}

function getProviderName(type: string) {
  const found = providers.find(p => p.id === type)
  return found ? found.name : type
}

function getProviderColorClass(type: string) {
  const found = providers.find(p => p.id === type)
  return found ? found.color : 'text-gray-600 bg-gray-50 border-gray-200'
}

function getStatusBadgeColor(status: string) {
  switch (status) {
    case 'Connected': return 'success'
    case 'Syncing': return 'info'
    case 'Error': return 'error'
    default: return 'neutral'
  }
}

function getStatusDotClass(status: string) {
  switch (status) {
    case 'Connected': return 'bg-emerald-500 animate-pulse'
    case 'Syncing': return 'bg-blue-500 animate-ping'
    case 'Error': return 'bg-rose-500'
    default: return 'bg-gray-400'
  }
}

// Testing & Sync actions
const testingId = ref<string | null>(null)
const syncingId = ref<string | null>(null)

async function testConnection(conn: DataSourceConn) {
  testingId.value = conn.id
  try {
    const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources/${conn.id}/test`, { method: 'POST' })
    if (res && res.data) {
      const idx = connections.value.findIndex(c => c.id === conn.id)
      if (idx !== -1) connections.value[idx] = res.data
      dataSourceStore.upsertSource(res.data)
    }
    if (res?.success === true) {
      toast.add({
        title: t('settings.dataSource.testSuccessToast'),
        description: res.message || t('toasts.dataSource.testSuccessDesc', { name: conn.name, host: conn.host, port: conn.port }),
        color: 'success',
      })
    } else {
      toast.add({
        title: t('toasts.dataSource.testFailedTitle'),
        description: res?.message || t('toasts.dataSource.testFailedDesc', { name: conn.name, host: conn.host, port: conn.port }),
        color: 'error',
      })
    }
  } catch (err: any) {
    toast.add({
      title: t('toasts.dataSource.testFailedTitle'),
      description: err?.data?.message || err?.message || t('toasts.dataSource.testFailedDesc', { name: conn.name, host: conn.host, port: conn.port }),
      color: 'error',
    })
  } finally {
    testingId.value = null
  }
}

function pollSingleSyncStatus(conn: DataSourceConn, jobId: string, logId?: string) {
  let attempts = 0
  const timer = setInterval(async () => {
    attempts++
    try {
      const q = logId ? `?log_id=${encodeURIComponent(logId)}` : ''
      const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources/sync-status/${jobId}${q}`)
      const st = res?.data?.status
      if (st === 'completed' || st === 'failed' || attempts >= 20) {
        clearInterval(timer)
        await fetchConnectionsFromApi()
        await fetchLogsFromApi()
        if (st === 'completed') {
          const processed = res?.data?.records_processed ?? 0
          toast.add({
            title: t('settings.dataSource.syncCompletedToast'),
            description: `${conn.name}: ${Number(processed).toLocaleString('id-ID')} records ingested & CAATT Engine updated.`,
            color: 'success',
          })
        } else if (st === 'failed') {
          toast.add({
            title: 'Sync Job Failed',
            description: res?.data?.error_message || `Background ingestion failed for ${conn.name}.`,
            color: 'error',
          })
        }
      }
    } catch {
      if (attempts >= 12) clearInterval(timer)
    }
  }, 2000)
}

async function triggerSync(conn: DataSourceConn) {
  const rem = getSyncCooldownRemaining(conn.id)
  if (rem > 0) {
    toast.add({
      title: 'Sync Cooldown Active',
      description: `Harap tunggu ${rem} detik sebelum melakukan sinkronisasi ulang pada ${conn.name}.`,
      color: 'warning',
    })
    return
  }

  syncingId.value = conn.id
  const prevStatus = conn.status
  const prevLastSync = conn.lastSync
  conn.status = 'Syncing'
  conn.lastSync = 'Syncing in progress...'

  toast.add({
    title: t('settings.dataSource.syncInitToast'),
    description: t('toasts.dataSource.syncStartedDesc', { name: conn.name }),
    color: 'info',
  })

  try {
    const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources/${conn.id}/sync`, { method: 'POST' })
    setSyncCooldown(conn.id, res?.cooldown_seconds || SYNC_COOLDOWN_SEC)
    if (res && res.data) {
      const idx = connections.value.findIndex(c => c.id === conn.id)
      if (idx !== -1) connections.value[idx] = res.data
      dataSourceStore.upsertSource(res.data)
    }
    await fetchLogsFromApi()
    if (res?.job_id) {
      pollSingleSyncStatus(conn, res.job_id, res.log_id)
    } else {
      toast.add({
        title: t('settings.dataSource.syncCompletedToast'),
        description: `Data ingestion finished for ${conn.name}.`,
        color: 'success',
      })
    }
  } catch (err: any) {
    const status = err?.status || err?.statusCode || err?.response?.status
    const retryAfter = err?.data?.retry_after
    if (status === 429 && retryAfter) {
      setSyncCooldown(conn.id, Number(retryAfter))
      conn.status = prevStatus
      conn.lastSync = prevLastSync
      toast.add({
        title: 'Sync Rate Limited (Anti-Spam)',
        description: err?.data?.message || `Tunggu ${retryAfter} detik sebelum melakukan sync ulang.`,
        color: 'warning',
      })
    } else {
      conn.status = 'Error'
      conn.lastSync = prevLastSync
      conn.lastError = err?.data?.message || err?.message || 'Data Hub synchronization failed'
      await fetchLogsFromApi()
      toast.add({
        title: 'Data Synchronization Failed',
        description: conn.lastError,
        color: 'error',
      })
    }
  } finally {
    syncingId.value = null
  }
}

const isBatchSyncing = ref(false)
let batchPollInterval: any = null

async function syncAllSources() {
  if (connections.value.length === 0) return
  const rem = getSyncCooldownRemaining('__all__')
  if (rem > 0) {
    toast.add({
      title: 'Sync All Cooldown Active',
      description: `Harap tunggu ${rem} detik sebelum melakukan sinkronisasi massal lagi.`,
      color: 'warning',
    })
    return
  }

  isBatchSyncing.value = true

  connections.value.forEach(c => {
    c.status = 'Syncing'
    c.lastSync = 'Batch sync in progress...'
  })

  toast.add({
    title: t('settings.dataSource.batchSyncStartedToast'),
    description: t('toasts.dataSource.batchSyncStartedDesc', { count: connections.value.length }),
    color: 'info',
  })

  try {
    const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources/sync-all`, { method: 'POST' })
    setSyncCooldown('__all__', res?.cooldown_seconds || SYNC_COOLDOWN_SEC)
    if (res && res.job_id) {
      pollBatchStatus(res.job_id)
    } else {
      finishBatchSync()
    }
  } catch (err: any) {
    isBatchSyncing.value = false
    const status = err?.status || err?.statusCode || err?.response?.status
    const retryAfter = err?.data?.retry_after
    await fetchConnectionsFromApi()
    await fetchLogsFromApi()
    if (status === 429 && retryAfter) {
      setSyncCooldown('__all__', Number(retryAfter))
      toast.add({
        title: 'Sync All Cooldown Active',
        description: err?.data?.message || `Harap tunggu ${retryAfter} detik sebelum Sync All lagi.`,
        color: 'warning',
      })
    } else {
      toast.add({
        title: 'Batch Sync Failed',
        description: err?.data?.message || err?.message || 'Gagal menjalankan batch sync ke Data Hub.',
        color: 'error',
      })
    }
  }
}

function pollBatchStatus(jobId: string) {
  let attempts = 0
  if (batchPollInterval) clearInterval(batchPollInterval)

  batchPollInterval = setInterval(async () => {
    attempts++
    try {
      const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources/sync-status/${jobId}`)
      if (res && res.data) {
        if (res.data.status === 'completed' || res.data.status === 'failed' || attempts > 25) {
          clearInterval(batchPollInterval)
          finishBatchSync(res.data.status === 'failed' ? res.data.error_message : undefined)
        }
      }
    } catch (e) {
      if (attempts > 12) {
        clearInterval(batchPollInterval)
        finishBatchSync()
      }
    }
  }, 2000)
}

async function finishBatchSync(errorMsg?: string) {
  isBatchSyncing.value = false
  await fetchConnectionsFromApi()
  await fetchLogsFromApi()
  if (errorMsg) {
    toast.add({
      title: 'Batch Sync Failed',
      description: errorMsg,
      color: 'error',
    })
    return
  }
  toast.add({
    title: t('settings.dataSource.batchSyncCompletedToast'),
    description: t('toasts.dataSource.batchSyncFinishedDesc'),
    color: 'success',
  })
}

function getConnMenuItems(conn: DataSourceConn) {
  return [
    [
      {
        label: t('settings.dataSource.schemaExplorer'),
        icon: 'i-lucide-table-properties',
        onSelect: () => openSchemaExplorer(conn),
      },
      {
        label: t('settings.dataSource.editConnection'),
        icon: 'i-lucide-pencil',
        onSelect: () => openEditModal(conn),
      },
      {
        label: t('settings.dataSource.viewCredentials'),
        icon: 'i-lucide-key-round',
        onSelect: () => openEditModal(conn),
      },
    ],
    [
      {
        label: t('settings.dataSource.deleteConnection'),
        icon: 'i-lucide-trash-2',
        color: 'error' as const,
        onSelect: () => deleteConnection(conn),
      },
    ],
  ]
}

async function deleteConnection(conn: DataSourceConn) {
  try {
    await $fetch(`${getMasterServiceBaseUrl()}/data-sources/${conn.id}`, { method: 'DELETE' })
  } catch (err) {
    console.warn('Delete backend error, using local fallback:', err)
  }
  connections.value = connections.value.filter(c => c.id !== conn.id)
  dataSourceStore.removeSource(conn.id)
  toast.add({
    title: t('settings.dataSource.connRemovedToast'),
    description: t('toasts.dataSource.removedDesc', { name: conn.name }),
    color: 'neutral',
  })
}

// Modal state
const isModalOpen = ref(false)
const isEditMode = ref(false)
const showPassword = ref(false)
const isSaving = ref(false)
const isModalTesting = ref(false)
const modalTestResult = ref<{ success: boolean; message: string } | null>(null)
const editingConnId = ref<string | null>(null)

const defaultForm = () => ({
  name: '',
  type: 'postgres' as DataSourceConn['type'],
  host: '',
  port: 5432,
  database: '',
  environment: 'Production' as DataSourceConn['environment'],
  syncSchedule: 'Hourly' as SyncFrequencyOption,
  username: '',
  password: '',
  ssl: true,
  scopes: ['risk_management', 'audit_features'],
})

const form = ref(defaultForm())

function toggleScope(scopeId: string) {
  if (form.value.scopes.includes(scopeId)) {
    form.value.scopes = form.value.scopes.filter(s => s !== scopeId)
  } else {
    form.value.scopes.push(scopeId)
  }
}

function openAddModal() {
  isEditMode.value = false
  editingConnId.value = null
  form.value = defaultForm()
  modalTestResult.value = null
  isModalOpen.value = true
}

function openEditModal(conn: DataSourceConn) {
  isEditMode.value = true
  editingConnId.value = conn.id
  form.value = {
    name: conn.name,
    type: conn.type,
    host: conn.host,
    port: conn.port,
    database: conn.database,
    environment: conn.environment,
    syncSchedule: normalizeSyncSchedule(conn.syncSchedule),
    username: conn.username || '',
    password: conn.password || '',
    ssl: conn.ssl,
    scopes: conn.scopes ? [...conn.scopes] : ['risk_management', 'audit_features'],
  }
  modalTestResult.value = null
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
  modalTestResult.value = null
}

function selectProvider(provider: typeof providers[number]) {
  form.value.type = provider.id as DataSourceConn['type']
  form.value.port = provider.defaultPort
  modalTestResult.value = null
}

async function testModalConnection() {
  if (!form.value.host) {
    modalTestResult.value = { success: false, message: 'Please enter a valid Host / Server IP address.' }
    return
  }
  isModalTesting.value = true
  modalTestResult.value = null

  try {
    const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources/test`, {
      method: 'POST',
      body: { host: form.value.host, port: form.value.port, ssl: form.value.ssl }
    })
    modalTestResult.value = {
      success: res.success === true,
      message: res.message || `Connection test to ${form.value.host}:${form.value.port} returned no result.`
    }
  } catch (err: any) {
    modalTestResult.value = {
      success: false,
      message: err?.data?.message || err?.message || `Connection test to ${form.value.host}:${form.value.port} failed.`
    }
  } finally {
    isModalTesting.value = false
  }
}

async function saveConnection() {
  if (!form.value.name || !form.value.host) {
    toast.add({
      title: t('toasts.dataSource.validationTitle'),
      description: t('toasts.dataSource.nameHostRequired'),
      color: 'error',
    })
    return
  }

  isSaving.value = true
  let hubError: string | null = null

  if (isEditMode.value && editingConnId.value) {
    try {
      const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources/${editingConnId.value}`, {
        method: 'PUT',
        body: form.value
      })
      if (res && res.data) {
        const idx = connections.value.findIndex(c => c.id === editingConnId.value)
        if (idx !== -1) connections.value[idx] = res.data
        dataSourceStore.upsertSource(res.data)
      }
      hubError = res?.dataHubRegistered === false ? (res.dataHubError || 'unknown error') : null
    } catch (err) {
      // Fallback local update
      const idx = connections.value.findIndex(c => c.id === editingConnId.value)
      if (idx !== -1 && connections.value[idx]) {
        const current = connections.value[idx]
        connections.value[idx] = {
          ...current,
          id: editingConnId.value,
          name: form.value.name,
          type: form.value.type,
          host: form.value.host,
          port: form.value.port,
          database: form.value.database,
          environment: form.value.environment,
          syncSchedule: form.value.syncSchedule,
          username: form.value.username,
          password: form.value.password,
          ssl: form.value.ssl,
          scopes: [...form.value.scopes],
        }
      }
    }
    if (hubError) {
      toast.add({
        title: t('settings.dataSource.connUpdatedToast'),
        description: t('toasts.dataSource.savedHubFailed', { name: form.value.name, error: hubError }),
        color: 'warning',
      })
    } else {
      toast.add({
        title: t('settings.dataSource.connUpdatedToast'),
        description: t('toasts.dataSource.updatedDesc', { name: form.value.name }),
        color: 'success',
      })
    }
  } else {
    try {
      const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources`, {
        method: 'POST',
        body: form.value
      })
      if (res && res.data) {
        connections.value.unshift(res.data)
        dataSourceStore.upsertSource(res.data)
        fetchLogsFromApi()
      }
      hubError = res?.dataHubRegistered === false ? (res.dataHubError || 'unknown error') : null
    } catch (err) {
      // Local fallback create
      const newConn: DataSourceConn = {
        id: `ds-${Date.now()}`,
        name: form.value.name,
        type: form.value.type,
        host: form.value.host,
        port: form.value.port,
        database: form.value.database,
        environment: form.value.environment,
        status: 'Connected',
        syncSchedule: form.value.syncSchedule,
        lastSync: 'Just configured',
        ssl: form.value.ssl,
        username: form.value.username,
        password: form.value.password,
        scopes: [...form.value.scopes],
      }
      connections.value.unshift(newConn)
      dataSourceStore.upsertSource(newConn as any)
    }
    if (hubError) {
      toast.add({
        title: t('settings.dataSource.connCreatedToast'),
        description: t('toasts.dataSource.savedHubFailed', { name: form.value.name, error: hubError }),
        color: 'warning',
      })
    } else {
      toast.add({
        title: t('settings.dataSource.connCreatedToast'),
        description: t('toasts.dataSource.createdDesc', { name: form.value.name }),
        color: 'success',
      })
    }
  }

  isSaving.value = false
  closeModal()
}

// ─────────────────────────────────────────────────────────────
// Schema Introspection & Data Mapping State & Actions
// ─────────────────────────────────────────────────────────────
const isSchemaModalOpen = ref(false)
const selectedConnForSchema = ref<DataSourceConn | null>(null)
const isLoadingSchema = ref(false)
const isSavingMappings = ref(false)
const discoveredTables = ref<SchemaTable[]>([])
const activeMappings = ref<Record<string, TableDataMapping[]>>({})
const tableSearchQuery = ref('')
const scopeFilter = ref('All Scopes')

function getModulesForScope(scope?: string, table?: SchemaTable) {
  return getModulesForTargetScope(scope, table)
}

function getScopesForTable(table: SchemaTable) {
  return getScopeOptionsForTable(table)
}

function getTableArchetypeInfo(table: SchemaTable) {
  const signals = analyzeTableSchema(table)
  return {
    signals,
    meta: getTableArchetypeMeta(signals.archetype),
  }
}

function getTableRecommendedModules(table: SchemaTable) {
  return getRecommendedMappingsForTable(table)
}

function getUnmappedRecommendedCount(table: SchemaTable): number {
  const existing = new Set(getTableMappings(table.tableName).map(m => (m.targetModule || '').toLowerCase()))
  return getRecommendedMappingsForTable(table).filter(r => !existing.has(r.targetModule.toLowerCase())).length
}

function getMappingCompatibility(table: SchemaTable, targetModule: string): ModuleCompatibilityEvaluation {
  return evaluateModuleCompatibility(table, targetModule)
}

function getImportModeBadge(importMode?: string) {
  if (importMode === 'auto') {
    return { label: 'Impor Otomatis', color: 'success' as const }
  }
  if (importMode === 'manual_banner') {
    return { label: 'Konfirmasi Manual via Banner', color: 'warning' as const }
  }
  if (importMode === 'prefill_form') {
    return { label: 'Pre-fill Form F-03 via Banner', color: 'info' as const }
  }
  return { label: 'Belum Didukung', color: 'neutral' as const }
}

function formatImportTime(iso?: string): string {
  if (!iso) return ''
  try {
    return new Date(iso).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
  } catch {
    return ''
  }
}

function getActiveMappingsCountForConn(conn: DataSourceConn): number {
  return (conn.dataMappings || []).filter(m => m.isActive !== false).length
}

// Preview state
const isPreviewModalOpen = ref(false)
const previewTableName = ref('')
const previewRows = ref<Record<string, any>[]>([])
const isLoadingPreview = ref(false)

const activeMappingsCount = computed(() => {
  return Object.values(activeMappings.value).flat().filter(m => m.isActive).length
})

const filteredTables = computed(() => {
  return discoveredTables.value.filter(t => {
    const q = tableSearchQuery.value.toLowerCase()
    const matchesSearch = !q || t.tableName.toLowerCase().includes(q) ||
      (t.description && t.description.toLowerCase().includes(q)) ||
      t.columns.some(c => c.name.toLowerCase().includes(q))

    const mappings = activeMappings.value[t.tableName] || []
    let matchesScope = true
    if (scopeFilter.value !== 'All Scopes') {
      const scopeKey = scopeFilter.value === 'Risk Management' ? 'risk_management'
        : scopeFilter.value === 'Audit Features' ? 'audit_features'
        : scopeFilter.value === 'QAR Features' ? 'qar_features'
        : scopeFilter.value === 'Data Analytics / AI' ? 'data_analytics'
        : ''
      matchesScope = mappings.some(m => m.targetScope === scopeKey)
    }

    return matchesSearch && matchesScope
  })
})

function getScopeLabel(scopeId?: string) {
  if (scopeId === 'risk_management') return 'Risk Management'
  if (scopeId === 'audit_features') return 'Audit Features'
  if (scopeId === 'qar_features') return 'QAR Features'
  if (scopeId === 'data_analytics') return 'Data Analytics / AI'
  return scopeId || 'Unassigned'
}

function getTableMappings(tableName: string): TableDataMapping[] {
  return activeMappings.value[tableName] || []
}

function addTableMapping(table: SchemaTable) {
  const list = activeMappings.value[table.tableName] ?? (activeMappings.value[table.tableName] = [])
  const existingModules = list.map(m => m.targetModule)
  const recommended = getNextRecommendedMappingForTable(table, existingModules)
  const autoMeta = resolveAutoMappingMetadata(table, recommended.targetModule)

  list.push({
    tableName: table.tableName,
    targetScope: recommended.targetScope,
    targetModule: recommended.targetModule,
    ...autoMeta,
    syncInterval: normalizeSyncSchedule(selectedConnForSchema.value?.syncSchedule),
    anomalyRules: [],
    customRules: [],
    isActive: true,
  })
}

function applyRecommendedMappings(table: SchemaTable) {
  const list = activeMappings.value[table.tableName] ?? (activeMappings.value[table.tableName] = [])
  const existing = new Set(list.map(m => (m.targetModule || '').trim().toLowerCase()))
  const recs = getRecommendedMappingsForTable(table)
  let added = 0

  for (const rec of recs) {
    if (existing.has(rec.targetModule.toLowerCase())) continue
    const autoMeta = resolveAutoMappingMetadata(table, rec.targetModule)
    list.push({
      tableName: table.tableName,
      targetScope: rec.targetScope,
      targetModule: rec.targetModule,
      ...autoMeta,
      syncInterval: normalizeSyncSchedule(selectedConnForSchema.value?.syncSchedule),
      anomalyRules: [],
      customRules: [],
      isActive: true,
    })
    existing.add(rec.targetModule.toLowerCase())
    added++
  }

  if (added > 0) {
    toast.add({
      title: 'Rekomendasi Diterapkan',
      description: `${added} mapping fitur yang cocok untuk tabel '${table.tableName}' telah ditambahkan.`,
      color: 'success',
    })
  } else {
    toast.add({
      title: 'Semua Rekomendasi Sudah Terpasang',
      description: `Seluruh fitur yang direkomendasikan untuk tabel '${table.tableName}' sudah dimapping.`,
      color: 'info',
    })
  }
}

function fixIncompatibleMapping(table: SchemaTable, index: number) {
  const list = activeMappings.value[table.tableName]
  if (!list || !list[index]) return
  const otherModules = list.filter((_, i) => i !== index).map(m => m.targetModule)
  const recommended = getNextRecommendedMappingForTable(table, otherModules)
  const autoMeta = resolveAutoMappingMetadata(table, recommended.targetModule)
  list[index] = {
    ...list[index],
    targetScope: recommended.targetScope,
    targetModule: recommended.targetModule,
    canonicalEntity: autoMeta.canonicalEntity,
    auditField: autoMeta.auditField,
    watermarkColumn: autoMeta.watermarkColumn,
    softDeleteColumn: autoMeta.softDeleteColumn,
  }
}

function removeTableMapping(tableName: string, index: number) {
  if (activeMappings.value[tableName]) {
    activeMappings.value[tableName].splice(index, 1)
  }
}

function updateMappingField(tableName: string, index: number, field: keyof TableDataMapping, value: any) {
  const list = activeMappings.value[tableName]
  if (list && list[index]) {
    const table = discoveredTables.value.find(t => t.tableName === tableName)
    const nextItem: TableDataMapping = {
      ...list[index],
      [field]: value
    }
    if (field === 'targetScope') {
      nextItem.targetModule = getDefaultModuleForTargetScope(value, table)
    }
    if (field === 'targetScope' || field === 'targetModule') {
      const autoMeta = resolveAutoMappingMetadata(table, nextItem.targetModule)
      nextItem.canonicalEntity = autoMeta.canonicalEntity
      nextItem.auditField = autoMeta.auditField
      nextItem.watermarkColumn = autoMeta.watermarkColumn
      nextItem.softDeleteColumn = autoMeta.softDeleteColumn
    }
    list[index] = nextItem
  }
}

function toggleAnomalyRule(tableName: string, index: number, rule: string) {
  const list = activeMappings.value[tableName]
  if (list && list[index]) {
    const currentRules = list[index].anomalyRules ? [...list[index].anomalyRules] : []
    if (currentRules.includes(rule)) {
      list[index].anomalyRules = currentRules.filter(r => r !== rule)
    } else {
      list[index].anomalyRules = [...currentRules, rule]
    }
  }
}

function getMappingKey(tableName: string, mIdx: number): string {
  return `${tableName}_${mIdx}`
}

const activeRuleBuilderKey = ref<string | null>(null)
const isTrainingModel = ref<Record<string, boolean>>({})

const newCustomRule = ref<{
  name: string
  field: string
  operator: string
  threshold: number
  thresholdStr: string
  severity: 'HIGH' | 'MODERATE_HIGH' | 'MODERATE' | 'LOW_MODERATE' | 'LOW'
  category: string
  description: string
}>({
  name: '',
  field: '',
  operator: '>',
  threshold: 100000000,
  thresholdStr: '',
  severity: 'HIGH',
  category: 'Transaction',
  description: ''
})

const ruleSeverityOptions = [
  { label: 'High', value: 'HIGH' },
  { label: 'Moderate to High', value: 'MODERATE_HIGH' },
  { label: 'Moderate', value: 'MODERATE' },
  { label: 'Low to Moderate', value: 'LOW_MODERATE' },
  { label: 'Low', value: 'LOW' }
]

const ruleOperatorOptions = [
  { label: '> (Lebih Besar Dari)', value: '>' },
  { label: '>= (Lebih Besar Sama Dengan)', value: '>=' },
  { label: '< (Lebih Kecil Dari)', value: '<' },
  { label: '<= (Lebih Kecil Sama Dengan)', value: '<=' },
  { label: '== (Sama Dengan)', value: '==' },
  { label: '!= (Tidak Sama)', value: '!=' },
  { label: 'contains (Mengandung Teks)', value: 'contains' }
]

function getTableColumnOptions(table: SchemaTable) {
  if (!table || !table.columns) return []
  return table.columns.map(c => ({
    label: `${c.name} (${c.dataType})`,
    value: c.name
  }))
}

function openCustomRuleBuilder(tableName: string, mIdx: number, table: SchemaTable) {
  const defaultCol = table.columns.find(c => /amount|nominal|nilai|total|balance/i.test(c.name))?.name || table.columns[0]?.name || 'amount'
  newCustomRule.value = {
    name: '',
    field: defaultCol,
    operator: '>',
    threshold: 100000000,
    thresholdStr: '',
    severity: 'HIGH',
    category: getCanonicalCategory(table.tableName),
    description: ''
  }
  activeRuleBuilderKey.value = getMappingKey(tableName, mIdx)
}

function closeCustomRuleBuilder() {
  activeRuleBuilderKey.value = null
}

function saveCustomRule(tableName: string, mIdx: number) {
  const list = activeMappings.value[tableName]
  if (!list || !list[mIdx]) return

  const form = newCustomRule.value
  if (!form.name.trim()) {
    toast.add({
      title: t('toasts.dataSource.ruleNameRequiredTitle'),
      description: t('toasts.dataSource.ruleNameRequired'),
      color: 'warning'
    })
    return
  }

  const newRule: CustomAnomalyRule = {
    id: `crule-${Date.now()}-${Math.random().toString(36).substr(2, 4)}`,
    name: form.name.trim(),
    field: form.field,
    operator: form.operator,
    threshold: Number(form.threshold) || 0,
    thresholdStr: form.thresholdStr || undefined,
    severity: form.severity,
    category: form.category || getCanonicalCategory(tableName),
    description: form.description.trim() || `Pelanggaran ambang batas ${form.field} ${form.operator} ${form.threshold}`,
    isActive: true
  }

  if (!list[mIdx].customRules) {
    list[mIdx].customRules = []
  }
  list[mIdx].customRules.push(newRule)

  closeCustomRuleBuilder()
  toast.add({
    title: t('toasts.dataSource.ruleAddedTitle'),
    description: t('toasts.dataSource.ruleAddedDesc', { rule: newRule.name, table: tableName }),
    color: 'success'
  })
}

function removeCustomRule(tableName: string, mIdx: number, ruleId: string) {
  const list = activeMappings.value[tableName]
  if (list && list[mIdx] && list[mIdx].customRules) {
    list[mIdx].customRules = list[mIdx].customRules.filter(r => r.id !== ruleId)
  }
}

function toggleCustomRuleActive(tableName: string, mIdx: number, ruleId: string) {
  const list = activeMappings.value[tableName]
  if (list && list[mIdx] && list[mIdx].customRules) {
    const target = list[mIdx].customRules.find(r => r.id === ruleId)
    if (target) {
      target.isActive = !target.isActive
    }
  }
}

async function trainModelForMapping(table: SchemaTable, mIdx: number) {
  const list = activeMappings.value[table.tableName]
  if (!list || !list[mIdx]) return
  const mapping = list[mIdx]
  const key = getMappingKey(table.tableName, mIdx)

  const presetRules = mapping.anomalyRules || []
  const customRules = (mapping.customRules || []).filter(r => r.isActive !== false)

  if (presetRules.length === 0 && customRules.length === 0) {
    toast.add({
      title: t('toasts.dataSource.noRulesTitle'),
      description: t('toasts.dataSource.noRulesDesc'),
      color: 'warning'
    })
    return
  }

  isTrainingModel.value[key] = true
  try {
    const combinedRules = [
      ...presetRules.map((r, i) => ({
        id: `preset-${Date.now()}-${i}`,
        name: r,
        field: 'amount',
        operator: '>',
        threshold: 100000000,
        severity: 'HIGH',
        category: getCanonicalCategory(table.tableName),
        description: `Deteksi anomali aturan standar: ${r}`,
        isActive: true
      })),
      ...customRules
    ]

    const result = await trainAnomalyWithRules({
      connection_name: selectedConnForSchema.value?.name || 'Database Connection',
      table_name: table.tableName,
      category: getCanonicalCategory(table.tableName),
      rules: combinedRules,
      sample_records: previewRows.value.length > 0 && previewTableName.value === table.tableName ? previewRows.value : []
    })

    if (result && result.success) {
      const qPos = result.data?.retrain?.queue_position
      const qInfo = qPos ? ` (Antrian #${qPos})` : ''
      toast.add({
        title: 'Model Anomaly Detection Masuk Antrian Pelatihan',
        description: `Hanya model Anomaly Detection (Isolation Forest & XGBoost) yang dijadwalkan berbatch${qInfo} dengan ${combinedRules.length} aturan untuk tabel '${table.tableName}'.`,
        color: 'success'
      })
      const nowIso = new Date().toISOString()
      if (mapping.customRules) {
        mapping.customRules.forEach(r => { r.trainedAt = nowIso })
      }
    } else {
      toast.add({
        title: t('toasts.dataSource.trainInfoTitle'),
        description: result?.message || t('toasts.dataSource.trainedLocally'),
        color: 'info'
      })
    }
  } catch (err: any) {
    toast.add({
      title: t('toasts.dataSource.trainFailedTitle'),
      description: err?.message || t('toasts.dataSource.trainFailedDesc'),
      color: 'error'
    })
  } finally {
    isTrainingModel.value[key] = false
  }
}

const isReconcilingDeletes = ref(false)

async function triggerDeleteReconciliation() {
  if (!selectedConnForSchema.value) return
  isReconcilingDeletes.value = true
  try {
    const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources/${selectedConnForSchema.value.id}/reconcile`, {
      method: 'POST'
    })
    const delCount = res?.total_deletes_detected ?? 0
    toast.add({
      title: t('toasts.dataSource.reconcileTitle'),
      description: t('toasts.dataSource.reconcileDesc', { count: delCount }),
      color: delCount > 0 ? 'warning' : 'success'
    })
    fetchLogsFromApi()
  } catch (err: any) {
    fetchLogsFromApi()
    toast.add({
      title: 'Rekonsiliasi Deletion Gagal',
      description: err?.data?.message || err?.message || 'Data Hub tidak dapat dijangkau untuk melakukan rekonsiliasi penghapusan.',
      color: 'error'
    })
  } finally {
    isReconcilingDeletes.value = false
  }
}

async function openSchemaExplorer(conn: DataSourceConn) {
  selectedConnForSchema.value = conn
  isSchemaModalOpen.value = true
  isLoadingSchema.value = true
  tableSearchQuery.value = ''
  scopeFilter.value = 'All Scopes'
  activeMappings.value = {}

  // Hydrate existing mappings if already present in connection
  if (conn.dataMappings && conn.dataMappings.length > 0) {
    conn.dataMappings.forEach(m => {
      const list = activeMappings.value[m.tableName] ?? (activeMappings.value[m.tableName] = [])
      list.push({
        ...m,
        watermarkColumn: m.watermarkColumn === '__none__' ? '' : (m.watermarkColumn || ''),
        softDeleteColumn: m.softDeleteColumn === '__none__' ? '' : (m.softDeleteColumn || ''),
        anomalyRules: m.anomalyRules ? [...m.anomalyRules] : [],
        customRules: m.customRules
          ? m.customRules.map(cr => ({
              ...cr,
              severity: toAnomalyRiskCode(cr.severity) as CustomAnomalyRule['severity']
            }))
          : []
      })
    })
  }

  try {
    const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources/${conn.id}/schema`)
    if (res && res.data && res.data.tables) {
      discoveredTables.value = res.data.tables
    }
  } catch (err) {
    // Fallback standard banking schema
    discoveredTables.value = [
      {
        tableName: 'gl_transactions',
        rowCount: 142580,
        columnCount: 8,
        description: 'General Ledger posted financial transactions and journal vouchers',
        columns: [
          { name: 'id', dataType: 'uuid', isPrimary: true, isNullable: false },
          { name: 'transaction_ref', dataType: 'varchar(64)', isPrimary: false, isNullable: false },
          { name: 'account_number', dataType: 'varchar(32)', isPrimary: false, isNullable: false },
          { name: 'amount', dataType: 'numeric(18,2)', isPrimary: false, isNullable: false },
          { name: 'currency', dataType: 'varchar(3)', isPrimary: false, isNullable: false },
          { name: 'channel', dataType: 'varchar(32)', isPrimary: false, isNullable: true },
          { name: 'created_by', dataType: 'varchar(64)', isPrimary: false, isNullable: false },
          { name: 'created_at', dataType: 'timestamp', isPrimary: false, isNullable: false },
        ]
      },
      {
        tableName: 'loan_accounts',
        rowCount: 28490,
        columnCount: 7,
        description: 'Customer credit facilities, risk ratings & collateral values',
        columns: [
          { name: 'loan_id', dataType: 'varchar(32)', isPrimary: true, isNullable: false },
          { name: 'cif_number', dataType: 'varchar(20)', isPrimary: false, isNullable: false },
          { name: 'product_type', dataType: 'varchar(50)', isPrimary: false, isNullable: false },
          { name: 'principal_amount', dataType: 'numeric(18,2)', isPrimary: false, isNullable: false },
          { name: 'interest_rate', dataType: 'numeric(5,2)', isPrimary: false, isNullable: false },
          { name: 'kol_status', dataType: 'varchar(10)', isPrimary: false, isNullable: false },
          { name: 'disbursed_at', dataType: 'timestamp', isPrimary: false, isNullable: true },
        ]
      },
      {
        tableName: 'user_access_audit_logs',
        rowCount: 894200,
        columnCount: 6,
        description: 'Core banking privileged authentication, role mutations & terminal access',
        columns: [
          { name: 'log_id', dataType: 'bigserial', isPrimary: true, isNullable: false },
          { name: 'user_id', dataType: 'varchar(64)', isPrimary: false, isNullable: false },
          { name: 'action_name', dataType: 'varchar(128)', isPrimary: false, isNullable: false },
          { name: 'ip_address', dataType: 'varchar(45)', isPrimary: false, isNullable: true },
          { name: 'is_override_auth', dataType: 'boolean', isPrimary: false, isNullable: false },
          { name: 'logged_at', dataType: 'timestamp', isPrimary: false, isNullable: false },
        ]
      },
      {
        tableName: 'vendor_invoices',
        rowCount: 12350,
        columnCount: 7,
        description: 'Procurement invoice approvals, vendor details & disbursement status',
        columns: [
          { name: 'invoice_no', dataType: 'varchar(64)', isPrimary: true, isNullable: false },
          { name: 'vendor_code', dataType: 'varchar(32)', isPrimary: false, isNullable: false },
          { name: 'invoice_amount', dataType: 'numeric(18,2)', isPrimary: false, isNullable: false },
          { name: 'po_number', dataType: 'varchar(64)', isPrimary: false, isNullable: true },
          { name: 'approval_status', dataType: 'varchar(32)', isPrimary: false, isNullable: false },
          { name: 'approver_id', dataType: 'varchar(64)', isPrimary: false, isNullable: true },
          { name: 'paid_at', dataType: 'timestamp', isPrimary: false, isNullable: true },
        ]
      },
      {
        tableName: 'compliance_incident_records',
        rowCount: 320,
        columnCount: 6,
        description: 'Regulatory policy violations, STR/AML alerts & AML screening hits',
        columns: [
          { name: 'incident_id', dataType: 'uuid', isPrimary: true, isNullable: false },
          { name: 'category', dataType: 'varchar(64)', isPrimary: false, isNullable: false },
          { name: 'severity_score', dataType: 'integer', isPrimary: false, isNullable: false },
          { name: 'investigator_notes', dataType: 'text', isPrimary: false, isNullable: true },
          { name: 'status', dataType: 'varchar(32)', isPrimary: false, isNullable: false },
          { name: 'reported_at', dataType: 'timestamp', isPrimary: false, isNullable: false },
        ]
      },
    ]
  } finally {
    isLoadingSchema.value = false
  }
}

function closeSchemaModal() {
  isSchemaModalOpen.value = false
}

async function persistSchemaMappings(currentConn: DataSourceConn): Promise<TableDataMapping[]> {
  const mappingsArray = Object.values(activeMappings.value).flat().map(m => {
    const table = discoveredTables.value.find(t => t.tableName === m.tableName)
    const autoMeta = resolveAutoMappingMetadata(table, m.targetModule)
    return {
      ...m,
      canonicalEntity: autoMeta.canonicalEntity,
      auditField: m.auditField || autoMeta.auditField,
      watermarkColumn: m.watermarkColumn === '__none__' ? '' : (m.watermarkColumn || autoMeta.watermarkColumn),
      softDeleteColumn: m.softDeleteColumn === '__none__' ? '' : (m.softDeleteColumn || autoMeta.softDeleteColumn),
      syncInterval: normalizeSyncSchedule(currentConn.syncSchedule),
    }
  })

  try {
    const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources/${currentConn.id}/mappings`, {
      method: 'PUT',
      body: { dataMappings: mappingsArray }
    })
    if (res && res.data) {
      const idx = connections.value.findIndex(c => c.id === currentConn.id)
      if (idx !== -1) connections.value[idx] = res.data
      currentConn.dataMappings = mappingsArray
      dataSourceStore.upsertSource(res.data)
      fetchLogsFromApi()
    }
  } catch (err) {
    // Local fallback update
    const idx = connections.value.findIndex(c => c.id === currentConn.id)
    if (idx !== -1 && connections.value[idx]) {
      const updatedConn = connections.value[idx]!
      updatedConn.dataMappings = mappingsArray
      currentConn.dataMappings = mappingsArray
      dataSourceStore.upsertSource({
        ...updatedConn,
        source_type: updatedConn.type,
      })
    }
  }

  return mappingsArray
}

function showBulkImportSummaryToast(summary: BulkImportSummary) {
  const parts: string[] = []
  if (summary.importedMappings > 0) {
    parts.push(`${summary.importedMappings} mapping berhasil diimpor (${summary.totalRecordsAdded} data baru)`)
  }
  if (summary.skippedMappings > 0) {
    parts.push(`${summary.skippedMappings} sudah sinkron (dedupe)`)
  }
  if (summary.manualMappings > 0) {
    parts.push(`${summary.manualMappings} menunggu konfirmasi manual di banner`)
  }
  if (summary.incompatibleMappings > 0) {
    parts.push(`${summary.incompatibleMappings} tidak cocok`)
  }
  if (summary.errorMappings > 0) {
    parts.push(`${summary.errorMappings} gagal`)
  }

  const color =
    summary.errorMappings > 0 && summary.importedMappings === 0
      ? 'error'
      : summary.importedMappings > 0
        ? 'success'
        : 'info'

  toast.add({
    title: `Impor Mapping Selesai — ${summary.connectionName}`,
    description: parts.join(' • ') || 'Tidak ada mapping aktif untuk diimpor.',
    color,
  })
}

async function saveSchemaMappings() {
  const currentConn = selectedConnForSchema.value
  if (!currentConn) return
  isSavingMappings.value = true

  try {
    const mappingsArray = await persistSchemaMappings(currentConn)
    toast.add({
      title: t('settings.dataSource.mappingsSavedToast'),
      description: t('toasts.dataSource.mappingsSavedDesc', { count: mappingsArray.filter(m => m.isActive).length, name: currentConn.name }),
      color: 'success',
    })
    closeSchemaModal()
    // Auto-trigger sync pipeline immediately when not in cooldown
    if (getSyncCooldownRemaining(currentConn.id) === 0) {
      triggerSync(currentConn)
    }
  } finally {
    isSavingMappings.value = false
  }
}

async function saveAndImportAllMappings() {
  const currentConn = selectedConnForSchema.value
  if (!currentConn) return
  isSavingMappings.value = true
  importingConnId.value = currentConn.id

  try {
    const mappingsArray = await persistSchemaMappings(currentConn)
    isSavingMappings.value = false
    const summary = await importAllForConnection({
      connection: { id: currentConn.id, name: currentConn.name },
      mappings: mappingsArray,
      schemaTables: discoveredTables.value,
    })
    showImportSummaryPanel.value = true
    showBulkImportSummaryToast(summary)
  } finally {
    isSavingMappings.value = false
    importingConnId.value = null
  }
}

async function importAllMappingsForConn(conn: DataSourceConn) {
  const mappings = (conn.dataMappings || []).filter(m => m.isActive !== false)
  if (mappings.length === 0) {
    toast.add({
      title: 'Belum Ada Mapping Aktif',
      description: `Buka Schema Explorer pada '${conn.name}' untuk memilih mapping tabel terlebih dahulu.`,
      color: 'warning',
    })
    return
  }

  importingConnId.value = conn.id
  try {
    const summary = await importAllForConnection({
      connection: { id: conn.id, name: conn.name },
      mappings,
    })
    showImportSummaryPanel.value = true
    showBulkImportSummaryToast(summary)
  } finally {
    importingConnId.value = null
  }
}

async function openPreviewModal(tableName: string) {
  if (!selectedConnForSchema.value) return
  previewTableName.value = tableName
  isPreviewModalOpen.value = true
  isLoadingPreview.value = true
  previewRows.value = []

  try {
    const res: any = await $fetch(`${getMasterServiceBaseUrl()}/data-sources/${selectedConnForSchema.value.id}/preview/${tableName}`)
    if (res && res.data && res.data.rows) {
      previewRows.value = res.data.rows
    }
  } catch (err) {
    if (tableName === 'gl_transactions') {
      previewRows.value = [
        { id: 'c1f721e0-0814-4eb9-bf2a-429f55e69e01', transaction_ref: 'TRX-20260813-0091', account_number: 'ACC-10029381', amount: 450000000.00, currency: 'IDR', channel: 'RTGS', created_by: 'teller_014', created_at: '2026-08-13 14:15:02' },
        { id: 'e4a812b1-5912-4fc8-9f1a-518f44d78a02', transaction_ref: 'TRX-20260813-0092', account_number: 'ACC-50019283', amount: 12500000.00, currency: 'IDR', channel: 'MOBILE_APP', created_by: 'system_auto', created_at: '2026-08-13 14:22:11' },
      ]
    } else {
      previewRows.value = [
        { id: '1', name: 'Sample row 1', status: 'ACTIVE' },
        { id: '2', name: 'Sample row 2', status: 'COMPLETED' },
      ]
    }
  } finally {
    isLoadingPreview.value = false
  }
}
</script>
