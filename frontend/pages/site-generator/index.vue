<template>
  <div class="min-h-screen bg-[var(--bg-main)] text-[var(--text-main)] transition-colors duration-300 flex flex-col">
    <!-- Standalone Header Navigation -->
    <header class="sticky top-0 z-40 w-full border-b border-[var(--border-main)] bg-[var(--bg-surface)]/85 backdrop-blur-md">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between gap-4">
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2">
            <Logo class="h-7 w-auto" />
          </div>

          <div class="h-5 w-px bg-[var(--border-main)] hidden sm:block" />

          <div class="hidden sm:flex items-center gap-2">
            <UBadge
              variant="solid"
              color="primary"
            >
              <UIcon
                name="i-lucide-globe"
                class="w-3.5 h-3.5"
              />
              {{ t('siteGenerator.badge') }}
            </UBadge>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <ReusableButton
            variant="ghost"
            color="neutral"
            size="sm"
            icon="i-lucide-layout-dashboard"
            @click="navigateTo('/dashboard')"
          >
            {{ t('siteGenerator.forbidden.backToDashboard') }}
          </ReusableButton>
          <UColorModeButton />
        </div>
      </div>
    </header>

    <main class="flex-1 flex flex-col items-center justify-center gap-6 p-4 sm:p-6 lg:p-8">
      <!-- FORBIDDEN STATE (the global middleware redirects too; auth-service checks the role again) -->
      <UCard
        v-if="!isPlatformAdmin"
        class="w-full max-w-lg"
      >
        <div class="p-6 text-center space-y-3">
          <div class="w-12 h-12 mx-auto rounded-xl bg-red-50 dark:bg-red-950/50 border border-red-200 dark:border-red-900 flex items-center justify-center text-red-600 dark:text-red-400">
            <UIcon
              name="i-lucide-shield-x"
              class="w-6 h-6"
            />
          </div>
          <h2 class="text-lg font-bold text-gray-900 dark:text-white">
            {{ t('siteGenerator.forbidden.title') }}
          </h2>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('siteGenerator.forbidden.description') }}
          </p>
          <ReusableButton
            variant="fill"
            color="primary"
            size="sm"
            @click="navigateTo('/dashboard')"
          >
            {{ t('siteGenerator.forbidden.backToDashboard') }}
          </ReusableButton>
        </div>
      </UCard>

      <template v-else>
        <!-- Provisioner unreachable (503) -->
        <div
          v-if="serviceError === 'unreachable'"
          class="w-full max-w-6xl p-3.5 rounded-xl bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-900/60 text-xs text-red-800 dark:text-red-300 flex items-start gap-2.5"
        >
          <UIcon
            name="i-lucide-server-off"
            class="w-4 h-4 shrink-0 mt-0.5"
          />
          <div class="flex-1 space-y-0.5">
            <p class="font-semibold">
              {{ t('siteGenerator.errors.unreachable') }}
            </p>
            <p>{{ t('siteGenerator.errors.unreachableHint') }}</p>
          </div>
          <ReusableButton
            variant="outline"
            color="error"
            size="xs"
            icon="i-lucide-refresh-cw"
            :loading="historyLoading"
            @click="loadInitialData"
          >
            {{ t('siteGenerator.actions.retry') }}
          </ReusableButton>
        </div>

        <UCard
          class="w-full max-w-6xl shadow-xl transition-all"
          :ui="{
            root: 'w-full max-w-6xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden flex flex-col transition-all ring-0 divide-y-0',
            header: 'p-0 sm:px-0',
            body: 'p-0 sm:p-0 flex-1 overflow-hidden flex flex-col',
            footer: 'px-6 sm:px-8 py-4 border-t border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-transparent flex items-center justify-between'
          }"
        >
          <template #header>
            <div class="px-6 sm:px-8 py-5 flex items-center justify-between gap-3">
              <div class="flex items-center gap-3.5">
                <div class="w-10 h-10 rounded-xl bg-primary-50 dark:bg-primary-950/60 border border-primary-200 dark:border-primary-800 flex items-center justify-center text-primary-600 dark:text-primary-400 shrink-0 shadow-sm">
                  <UIcon
                    :name="view === 'job' ? 'i-lucide-activity' : 'i-lucide-globe'"
                    class="w-5 h-5"
                  />
                </div>
                <div>
                  <h2 class="text-base font-bold text-gray-900 dark:text-white">
                    {{ view === 'job' ? t('siteGenerator.job.title') : t('siteGenerator.title') }}
                  </h2>
                  <p class="text-xs text-gray-500 dark:text-gray-400">
                    {{ t('siteGenerator.subtitle') }}
                  </p>
                </div>
              </div>

              <span
                v-if="view === 'wizard'"
                class="text-xs font-semibold px-2.5 py-1 rounded-full bg-primary/10 text-primary border border-primary/20 shrink-0"
              >
                {{ t('siteGenerator.stepProgress', { current: currentStep, total: steps.length }) }}
              </span>
            </div>

            <!-- Stepper -->
            <div
              v-if="view === 'wizard'"
              class="px-6 sm:px-8 py-3 bg-gray-50/70 dark:bg-gray-800/40 border-y border-gray-100 dark:border-gray-800"
            >
              <div class="flex items-center w-full">
                <template
                  v-for="(step, idx) in steps"
                  :key="step.id"
                >
                  <button
                    type="button"
                    class="flex items-center gap-2 select-none shrink-0"
                    :class="currentStep > idx + 1 ? 'cursor-pointer' : 'cursor-default'"
                    @click="currentStep > idx + 1 ? currentStep = idx + 1 : null"
                  >
                    <span
                      :class="[
                        'w-6 h-6 rounded-full text-xs font-bold flex items-center justify-center transition-all duration-200',
                        currentStep === idx + 1
                          ? 'bg-primary text-white shadow-sm ring-2 ring-primary/20'
                          : currentStep > idx + 1
                            ? 'bg-emerald-500 text-white'
                            : 'bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-400'
                      ]"
                    >
                      <UIcon
                        v-if="currentStep > idx + 1"
                        name="i-lucide-check"
                        class="w-3.5 h-3.5"
                      />
                      <span v-else>{{ idx + 1 }}</span>
                    </span>
                    <span
                      :class="[
                        'text-xs font-medium hidden sm:inline transition-colors',
                        currentStep === idx + 1 ? 'text-gray-900 dark:text-white font-semibold' : 'text-gray-500 dark:text-gray-400'
                      ]"
                    >
                      {{ step.title }}
                    </span>
                  </button>

                  <div
                    v-if="idx < steps.length - 1"
                    class="flex-1 flex items-center justify-center min-w-4"
                  >
                    <UIcon
                      name="i-lucide-chevron-right"
                      class="w-3.5 h-3.5 text-gray-300 dark:text-gray-600 hidden sm:inline"
                    />
                  </div>
                </template>
              </div>
            </div>
          </template>

          <!-- LIVE JOB VIEW -->
          <ProvisioningJobView
            v-if="view === 'job' && activeJobId"
            :job-id="activeJobId"
            :state="jobState"
            @start-new="startNewJob"
            @copy="copyText"
          />

          <!-- WIZARD -->
          <div
            v-else
            class="p-6 sm:p-8 space-y-6 flex-1 overflow-y-auto max-h-[62vh]"
          >
            <!-- STEP 1: Organization & domain -->
            <div
              v-if="currentStep === 1"
              class="space-y-5"
            >
              <AppFormField
                :label="t('siteGenerator.organization.clientName')"
                :tooltip="t('siteGenerator.organization.clientNameTooltip')"
                :error="form.clientName && !isValidClientName(form.clientName) ? t('siteGenerator.organization.clientNameError', { min: CLIENT_NAME_MIN, max: CLIENT_NAME_MAX }) : undefined"
                required
                counter
                :max-count="CLIENT_NAME_MAX"
                :model-value="form.clientName"
              >
                <UInput
                  v-model="form.clientName"
                  :placeholder="t('siteGenerator.organization.clientNamePlaceholder')"
                  :maxlength="CLIENT_NAME_MAX"
                  class="w-full"
                />
              </AppFormField>

              <!-- Environment -->
              <div class="flex items-center justify-between pb-0.5 gap-3 flex-wrap">
                <label class="text-xs font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
                  <UIcon
                    name="i-lucide-server"
                    class="w-3.5 h-3.5 text-primary"
                  />
                  {{ t('siteGenerator.organization.environment') }}
                </label>
                <div class="inline-flex rounded-lg p-0.5 bg-gray-100 dark:bg-neutral-800 border border-gray-200 dark:border-neutral-700 text-xs">
                  <button
                    v-for="opt in modeOptions"
                    :key="opt.value"
                    type="button"
                    :class="[
                      'px-3 py-1 rounded-md font-medium transition-all flex items-center gap-1.5',
                      form.mode === opt.value
                        ? 'bg-white dark:bg-neutral-900 text-primary shadow-sm font-semibold'
                        : 'text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
                    ]"
                    @click="form.mode = opt.value"
                  >
                    <UIcon
                      :name="opt.icon"
                      class="w-3.5 h-3.5"
                    />
                    {{ opt.label }}
                  </button>
                </div>
              </div>

              <div
                v-if="form.mode === 'local'"
                class="p-3 rounded-xl bg-blue-50 dark:bg-blue-950/40 border border-blue-200 dark:border-blue-900/50 text-xs text-blue-800 dark:text-blue-300 flex items-start gap-2.5"
              >
                <UIcon
                  name="i-lucide-info"
                  class="w-4 h-4 text-blue-600 dark:text-blue-400 shrink-0 mt-0.5"
                />
                <p class="leading-relaxed">
                  {{ t('siteGenerator.organization.localhostDescription') }}
                </p>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <AppFormField
                  :label="t('siteGenerator.organization.slug')"
                  :tooltip="t('siteGenerator.organization.slugTooltip')"
                  :hint="t('siteGenerator.organization.slugHint')"
                  :error="form.slug && !isValidSlug(form.slug) ? t('siteGenerator.organization.slugError') : undefined"
                  required
                >
                  <UInput
                    v-model="form.slug"
                    placeholder="bai"
                    maxlength="31"
                    class="w-full"
                    :ui="{ base: 'font-mono text-primary font-semibold' }"
                    @update:model-value="onSlugInput"
                  />
                </AppFormField>

                <AppFormField
                  v-if="form.mode === 'production'"
                  :label="t('siteGenerator.organization.domain')"
                  :tooltip="t('siteGenerator.organization.domainTooltip')"
                  :error="form.baseDomain && !isValidHostname(form.baseDomain) ? t('siteGenerator.organization.domainError') : undefined"
                  required
                >
                  <div class="flex items-center">
                    <span class="inline-flex items-center px-3 py-1.5 rounded-l-md border border-r-0 border-gray-300 dark:border-neutral-700 bg-gray-50 dark:bg-neutral-800 text-gray-600 dark:text-gray-300 text-sm font-mono select-none">
                      {{ slugOrPlaceholder }}.
                    </span>
                    <UInput
                      v-model="form.baseDomain"
                      :placeholder="DEFAULT_BASE_DOMAIN"
                      class="flex-1"
                      :ui="{ base: 'rounded-l-none font-mono' }"
                    />
                  </div>
                </AppFormField>
              </div>

              <!-- Expected URLs (the job result has the real ones) -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div class="p-3.5 rounded-xl border border-gray-200 dark:border-neutral-800 bg-gray-50 dark:bg-neutral-900/90 flex items-start gap-3">
                  <div class="w-8 h-8 rounded-lg bg-primary-50 dark:bg-primary-950/60 border border-primary-200 dark:border-primary-800 flex items-center justify-center text-primary shrink-0">
                    <UIcon
                      name="i-lucide-laptop"
                      class="w-4 h-4"
                    />
                  </div>
                  <div class="text-xs min-w-0 flex-1">
                    <div class="text-gray-500 dark:text-neutral-400 font-medium">
                      {{ t('siteGenerator.organization.frontendDomain') }}
                    </div>
                    <div class="font-mono font-semibold text-gray-900 dark:text-neutral-100 mt-0.5 break-all">
                      {{ expectedFrontendUrl }}
                    </div>
                  </div>
                </div>
                <div class="p-3.5 rounded-xl border border-gray-200 dark:border-neutral-800 bg-gray-50 dark:bg-neutral-900/90 flex items-start gap-3">
                  <div class="w-8 h-8 rounded-lg bg-secondary-50 dark:bg-secondary-950/60 border border-secondary-200 dark:border-secondary-800 flex items-center justify-center text-secondary shrink-0">
                    <UIcon
                      name="i-lucide-network"
                      class="w-4 h-4"
                    />
                  </div>
                  <div class="text-xs min-w-0 flex-1">
                    <div class="text-gray-500 dark:text-neutral-400 font-medium">
                      {{ t('siteGenerator.organization.apiEndpoint') }}
                    </div>
                    <div class="font-mono font-semibold text-gray-900 dark:text-neutral-100 mt-0.5 break-all">
                      {{ expectedApiUrl }}
                    </div>
                  </div>
                </div>
              </div>
              <p class="text-[11px] text-gray-400">
                {{ t('siteGenerator.organization.urlsEstimated') }}
              </p>
            </div>

            <!-- STEP 2: Administrator & data -->
            <div
              v-else-if="currentStep === 2"
              class="space-y-5"
            >
              <div class="p-3.5 rounded-xl bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-900/40 flex items-center gap-3">
                <UIcon
                  name="i-lucide-shield-alert"
                  class="w-5 h-5 text-blue-600 dark:text-blue-400 shrink-0"
                />
                <p class="text-xs text-blue-800 dark:text-blue-300 leading-relaxed">
                  {{ t('siteGenerator.administrator.info') }}
                </p>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <AppFormField
                  :label="t('siteGenerator.administrator.fullName')"
                  :tooltip="t('siteGenerator.administrator.fullNameTooltip')"
                  :error="!isValidAdminName(form.adminName) ? t('siteGenerator.administrator.fullNameError', { min: ADMIN_NAME_MIN, max: ADMIN_NAME_MAX }) : undefined"
                  :hint="t('siteGenerator.administrator.fullNameOptional')"
                  counter
                  :max-count="ADMIN_NAME_MAX"
                  :model-value="form.adminName"
                >
                  <UInput
                    v-model="form.adminName"
                    :placeholder="t('siteGenerator.administrator.fullNamePlaceholder')"
                    :maxlength="ADMIN_NAME_MAX"
                    class="w-full"
                  />
                </AppFormField>

                <AppFormField
                  :label="t('siteGenerator.administrator.email')"
                  :tooltip="t('siteGenerator.administrator.emailTooltip')"
                  :error="form.adminEmail && !isValidEmail(form.adminEmail) ? t('siteGenerator.administrator.emailError') : undefined"
                  required
                >
                  <UInput
                    v-model="form.adminEmail"
                    type="email"
                    placeholder="cae@example.com"
                    class="w-full"
                  />
                </AppFormField>
              </div>

              <div class="space-y-3">
                <div class="p-3.5 rounded-xl border border-gray-200 dark:border-neutral-800 flex items-start gap-3">
                  <USwitch
                    v-model="form.withDemoData"
                    class="mt-0.5"
                  />
                  <div class="text-xs space-y-0.5">
                    <p class="font-semibold text-gray-800 dark:text-gray-200">
                      {{ t('siteGenerator.administrator.demoData') }}
                    </p>
                    <p class="text-gray-500 dark:text-gray-400">
                      {{ t('siteGenerator.administrator.demoDataHint') }}
                    </p>
                  </div>
                </div>

                <div class="p-3.5 rounded-xl border border-gray-200 dark:border-neutral-800 flex items-start gap-3">
                  <USwitch
                    v-model="form.provisionEmail"
                    :disabled="form.mode === 'local'"
                    class="mt-0.5"
                  />
                  <div class="text-xs space-y-0.5">
                    <p class="font-semibold text-gray-800 dark:text-gray-200">
                      {{ t('siteGenerator.administrator.provisionEmail') }}
                    </p>
                    <p class="text-gray-500 dark:text-gray-400">
                      {{ form.mode === 'local' ? t('siteGenerator.administrator.provisionEmailLocal') : t('siteGenerator.administrator.provisionEmailHint') }}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <!-- STEP 3: Infrastructure -->
            <div
              v-else-if="currentStep === 3"
              class="space-y-5"
            >
              <div class="p-4 rounded-xl border border-primary-200 dark:border-primary-800/60 bg-primary-50/50 dark:bg-primary-950/20 flex items-start gap-3">
                <USwitch
                  v-model="form.autoPorts"
                  class="mt-0.5"
                />
                <div class="space-y-1 text-xs">
                  <p class="font-bold text-gray-900 dark:text-white">
                    {{ t('siteGenerator.infrastructure.autoPorts') }}
                  </p>
                  <p class="text-gray-500 dark:text-gray-400 leading-relaxed">
                    {{ t('siteGenerator.infrastructure.autoPortsDesc') }}
                  </p>
                  <p class="font-mono text-gray-700 dark:text-gray-300 flex items-center gap-1.5 pt-1">
                    <UIcon
                      v-if="portsLoading"
                      name="i-lucide-loader-2"
                      class="w-3.5 h-3.5 animate-spin"
                    />
                    <template v-if="portsLoading">
                      {{ t('siteGenerator.infrastructure.nextPortsLoading') }}
                    </template>
                    <template v-else-if="nextPorts">
                      {{ t('siteGenerator.infrastructure.nextPorts', { fe: nextPorts.fe_port, api: nextPorts.api_port }) }}
                    </template>
                    <span
                      v-else-if="portsError"
                      class="text-amber-700 dark:text-amber-400 font-sans"
                    >
                      {{ t(`siteGenerator.errors.${portsError}`) }}
                    </span>
                    <button
                      type="button"
                      class="text-primary hover:underline font-sans ml-1"
                      :disabled="portsLoading"
                      @click="loadNextPorts"
                    >
                      {{ t('siteGenerator.actions.refresh') }}
                    </button>
                  </p>
                </div>
              </div>

              <div
                v-if="!form.autoPorts"
                class="grid grid-cols-1 sm:grid-cols-2 gap-4"
              >
                <AppFormField
                  :label="t('siteGenerator.infrastructure.frontendPort')"
                  :tooltip="t('siteGenerator.infrastructure.frontendPortTooltip')"
                  :error="portError(form.fePort)"
                  required
                >
                  <UInput
                    v-model.number="form.fePort"
                    type="number"
                    :min="PORT_MIN"
                    :max="PORT_MAX"
                    :placeholder="String(nextPorts?.fe_port ?? 3015)"
                    class="w-full font-mono text-xs"
                  />
                </AppFormField>
                <AppFormField
                  :label="t('siteGenerator.infrastructure.apiPort')"
                  :tooltip="t('siteGenerator.infrastructure.apiPortTooltip')"
                  :error="portError(form.apiPort) || (form.fePort && form.fePort === form.apiPort ? t('siteGenerator.infrastructure.portsSame') : undefined)"
                  required
                >
                  <UInput
                    v-model.number="form.apiPort"
                    type="number"
                    :min="PORT_MIN"
                    :max="PORT_MAX"
                    :placeholder="String(nextPorts?.api_port ?? 8095)"
                    class="w-full font-mono text-xs"
                  />
                </AppFormField>
                <p class="sm:col-span-2 text-[11px] text-gray-400">
                  {{ t('siteGenerator.infrastructure.manualPortsHint') }}
                </p>
              </div>

              <!-- Production only: finish on plain HTTP when DNS/certbot is not ready (allow_no_tls) -->
              <div
                v-if="form.mode === 'production'"
                class="p-4 rounded-xl border space-y-2 transition-colors"
                :class="form.allowNoTls ? 'border-amber-300 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/30' : 'border-gray-200 dark:border-neutral-800'"
              >
                <UCheckbox
                  v-model="form.allowNoTls"
                  color="warning"
                  :label="t('siteGenerator.infrastructure.allowNoTls')"
                  :description="t('siteGenerator.infrastructure.allowNoTlsHint')"
                  :ui="{ label: 'text-xs font-semibold text-gray-800 dark:text-gray-200', description: 'text-xs' }"
                />
                <div
                  v-if="form.allowNoTls"
                  class="flex items-start gap-2 text-xs text-amber-800 dark:text-amber-300"
                  role="alert"
                >
                  <UIcon
                    name="i-lucide-shield-alert"
                    class="w-4 h-4 shrink-0 mt-0.5"
                  />
                  <p>{{ t('siteGenerator.infrastructure.allowNoTlsWarning') }}</p>
                </div>
              </div>

              <!-- Advanced: health timeout (health_timeout_seconds) -->
              <div class="rounded-xl border border-gray-200 dark:border-neutral-800">
                <button
                  type="button"
                  class="w-full px-4 py-3 text-xs font-semibold text-gray-700 dark:text-gray-300 flex items-center justify-between gap-2 hover:text-primary"
                  :aria-expanded="showAdvanced"
                  @click="showAdvanced = !showAdvanced"
                >
                  <span class="flex items-center gap-1.5">
                    <UIcon
                      name="i-lucide-sliders-horizontal"
                      class="w-3.5 h-3.5 text-primary"
                    />
                    {{ t('siteGenerator.infrastructure.advanced') }}
                  </span>
                  <UIcon
                    :name="showAdvanced ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
                    class="w-3.5 h-3.5"
                  />
                </button>
                <div
                  v-if="showAdvanced"
                  class="px-4 pb-4"
                >
                  <AppFormField
                    :label="t('siteGenerator.infrastructure.healthTimeout')"
                    :tooltip="t('siteGenerator.infrastructure.healthTimeoutTooltip')"
                    :error="healthTimeoutError"
                    :hint="t('siteGenerator.infrastructure.healthTimeoutHint', { default: HEALTH_TIMEOUT_DEFAULT, min: HEALTH_TIMEOUT_MIN, max: HEALTH_TIMEOUT_MAX })"
                  >
                    <UInput
                      v-model.number="form.healthTimeoutSeconds"
                      type="number"
                      :min="HEALTH_TIMEOUT_MIN"
                      :max="HEALTH_TIMEOUT_MAX"
                      step="1"
                      :placeholder="String(HEALTH_TIMEOUT_DEFAULT)"
                      class="w-full sm:w-48 font-mono text-xs"
                    >
                      <template #trailing>
                        <span class="text-xs text-gray-400">{{ t('siteGenerator.infrastructure.seconds') }}</span>
                      </template>
                    </UInput>
                  </AppFormField>
                </div>
              </div>

              <!-- What the script creates -->
              <div class="rounded-xl border border-dashed border-gray-200 dark:border-neutral-800 bg-gray-50/70 dark:bg-neutral-900/60 p-4 space-y-3 text-xs">
                <div class="space-y-1">
                  <span class="text-gray-400 block">{{ t('siteGenerator.infrastructure.storage') }}</span>
                  <span class="font-mono font-medium text-gray-800 dark:text-gray-200 flex items-center gap-1 break-all">
                    <UIcon
                      name="i-lucide-hard-drive"
                      class="w-3.5 h-3.5 shrink-0"
                    />
                    {{ storagePath }}
                  </span>
                  <span class="text-[11px] text-gray-400 block">{{ t('siteGenerator.infrastructure.storagePolicy') }}</span>
                </div>
                <div class="space-y-1.5">
                  <span class="text-gray-400 block">{{ t('siteGenerator.infrastructure.databases') }}</span>
                  <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 font-mono">
                    <div
                      v-for="dbName in databases"
                      :key="dbName"
                      class="px-3 py-2 rounded-lg bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700/60 flex items-center justify-between"
                    >
                      <span>{{ dbName }}</span>
                      <UIcon
                        name="i-lucide-lock"
                        class="w-3.5 h-3.5 text-gray-400"
                      />
                    </div>
                  </div>
                  <span class="text-[11px] text-gray-400 block">{{ t('siteGenerator.infrastructure.databasesHint') }}</span>
                </div>
              </div>
            </div>

            <!-- STEP 4: Review & provision -->
            <div
              v-else-if="currentStep === 4"
              class="space-y-5"
            >
              <div class="rounded-xl border border-gray-200 dark:border-neutral-800 overflow-hidden text-xs">
                <div class="bg-gray-50 dark:bg-neutral-900/90 px-4 py-2.5 font-semibold text-gray-900 dark:text-white border-b border-gray-200 dark:border-neutral-800 flex items-center gap-1.5">
                  <UIcon
                    name="i-lucide-clipboard-check"
                    class="w-4 h-4 text-primary"
                  />
                  {{ t('siteGenerator.review.summary') }}
                </div>
                <dl class="p-4 grid grid-cols-1 sm:grid-cols-2 gap-y-3 gap-x-6">
                  <div
                    v-for="row in reviewRows"
                    :key="row.label"
                  >
                    <dt class="text-gray-400">
                      {{ row.label }}
                    </dt>
                    <dd
                      class="text-gray-800 dark:text-gray-200 break-all"
                      :class="row.mono ? 'font-mono' : 'font-semibold'"
                    >
                      {{ row.value }}
                    </dd>
                  </div>
                </dl>
              </div>

              <!-- Dry run toggle -->
              <div
                class="p-4 rounded-xl border flex items-start gap-3 transition-colors"
                :class="form.dryRun ? 'border-amber-300 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/30' : 'border-gray-200 dark:border-neutral-800'"
              >
                <USwitch
                  v-model="form.dryRun"
                  color="warning"
                  class="mt-0.5"
                />
                <div class="text-xs space-y-0.5">
                  <p class="font-bold text-gray-900 dark:text-white flex items-center gap-1.5">
                    <UIcon
                      name="i-lucide-flask-conical"
                      class="w-3.5 h-3.5 text-amber-600"
                    />
                    {{ t('siteGenerator.review.dryRun') }}
                  </p>
                  <p class="text-gray-500 dark:text-gray-400">
                    {{ t('siteGenerator.review.dryRunHint') }}
                  </p>
                </div>
              </div>

              <!-- Create-job error -->
              <div
                v-if="submitError"
                class="p-3 rounded-xl bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-900/60 text-xs text-red-800 dark:text-red-300 flex items-start gap-2"
              >
                <UIcon
                  name="i-lucide-circle-alert"
                  class="w-4 h-4 shrink-0 mt-0.5"
                />
                <div class="space-y-0.5">
                  <p class="font-semibold">
                    {{ t(`siteGenerator.errors.${submitError.kind}`) }}
                  </p>
                  <p v-if="submitError.message">
                    {{ submitError.message }}
                  </p>
                </div>
              </div>

              <!-- Secondary: manual CLI command -->
              <div class="space-y-2">
                <button
                  type="button"
                  class="text-xs font-semibold text-gray-600 dark:text-gray-300 flex items-center gap-1.5 hover:text-primary"
                  :aria-expanded="showCli"
                  @click="showCli = !showCli"
                >
                  <UIcon
                    name="i-lucide-terminal"
                    class="w-3.5 h-3.5 text-primary"
                  />
                  {{ showCli ? t('siteGenerator.review.hideCli') : t('siteGenerator.review.showCli') }}
                  <UIcon
                    :name="showCli ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
                    class="w-3.5 h-3.5"
                  />
                </button>
                <div
                  v-if="showCli"
                  class="space-y-1.5"
                >
                  <div class="flex items-center justify-between gap-2">
                    <p class="text-[11px] text-gray-500 dark:text-gray-400">
                      {{ t('siteGenerator.review.cliHint') }}
                    </p>
                    <button
                      type="button"
                      class="text-xs text-primary hover:text-primary-700 flex items-center gap-1 font-medium shrink-0"
                      @click="copyText(cliCommand, t('siteGenerator.copy.cliCommand'))"
                    >
                      <UIcon
                        name="i-lucide-copy"
                        class="w-3.5 h-3.5"
                      />
                      {{ t('siteGenerator.review.copyCommand') }}
                    </button>
                  </div>
                  <div class="bg-gray-950 text-gray-200 p-3.5 rounded-xl font-mono text-xs overflow-x-auto border border-gray-800">
                    <code class="whitespace-pre">{{ cliCommand }}</code>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Wizard footer -->
          <template
            v-if="view === 'wizard'"
            #footer
          >
            <button
              v-if="currentStep > 1"
              type="button"
              :disabled="submitting"
              class="px-4 py-2 rounded-xl text-xs font-semibold text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
              @click="currentStep--"
            >
              {{ t('siteGenerator.actions.back') }}
            </button>
            <div v-else />

            <div class="flex items-center gap-2">
              <ReusableButton
                v-if="currentStep < steps.length"
                variant="fill"
                color="primary"
                size="sm"
                :disabled="!isStepValid(currentStep)"
                @click="currentStep++"
              >
                {{ t('siteGenerator.actions.continue') }}
              </ReusableButton>

              <ReusableButton
                v-else
                variant="fill"
                :color="form.dryRun ? 'warning' : 'primary'"
                size="sm"
                :icon="form.dryRun ? 'i-lucide-flask-conical' : 'i-lucide-rocket'"
                :loading="submitting"
                :disabled="!isFormValid || submitting"
                @click="handleProvision"
              >
                {{ form.dryRun ? t('siteGenerator.actions.provisionDryRun') : t('siteGenerator.actions.provision') }}
              </ReusableButton>
            </div>
          </template>
        </UCard>

        <!-- Recent jobs -->
        <UCard class="w-full max-w-6xl">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-2">
                <UIcon
                  name="i-lucide-history"
                  class="w-4 h-4 text-primary"
                />
                {{ t('siteGenerator.history.title') }}
              </h3>
              <ReusableButton
                variant="ghost"
                color="neutral"
                size="xs"
                icon="i-lucide-refresh-cw"
                :loading="historyLoading"
                @click="loadHistory"
              >
                {{ t('siteGenerator.actions.refresh') }}
              </ReusableButton>
            </div>
          </template>

          <div
            v-if="historyLoading && !history.length"
            class="py-6 flex items-center justify-center gap-2 text-xs text-gray-500"
          >
            <UIcon
              name="i-lucide-loader-2"
              class="w-4 h-4 animate-spin"
            />
            {{ t('siteGenerator.history.loading') }}
          </div>
          <p
            v-else-if="historyError && !history.length"
            class="py-4 text-xs text-red-600 dark:text-red-400 text-center"
          >
            {{ t(`siteGenerator.errors.${historyError}`) }}
          </p>
          <p
            v-else-if="!history.length"
            class="py-4 text-xs text-gray-500 text-center"
          >
            {{ t('siteGenerator.history.empty') }}
          </p>
          <ul
            v-else
            class="divide-y divide-gray-100 dark:divide-neutral-800 -my-2"
          >
            <li
              v-for="job in history"
              :key="job.id"
            >
              <button
                type="button"
                class="w-full py-2.5 flex items-center justify-between gap-3 text-xs text-left hover:bg-gray-50 dark:hover:bg-neutral-800/50 rounded-lg px-2"
                :class="job.id === activeJobId && view === 'job' ? 'bg-primary-50/60 dark:bg-primary-950/20' : ''"
                @click="viewJob(job)"
              >
                <span class="flex items-center gap-2 min-w-0">
                  <UIcon
                    :name="historyIcon(job.status)"
                    :class="['w-4 h-4 shrink-0', historyIconClass(job.status)]"
                  />
                  <span class="font-mono font-semibold text-gray-900 dark:text-white truncate">{{ job.slug }}</span>
                  <UBadge
                    color="neutral"
                    variant="outline"
                    size="xs"
                  >
                    {{ job.mode }}
                  </UBadge>
                  <UBadge
                    v-if="job.dry_run"
                    color="warning"
                    variant="subtle"
                    size="xs"
                  >
                    {{ t('siteGenerator.job.dryRunBadge') }}
                  </UBadge>
                </span>
                <span class="flex items-center gap-3 shrink-0 text-gray-500 dark:text-gray-400">
                  <span class="hidden sm:inline">{{ formatDate(job.created_at) }}</span>
                  <span class="font-medium">{{ t(`siteGenerator.job.status.${job.status}`) }}</span>
                  <UIcon
                    name="i-lucide-chevron-right"
                    class="w-3.5 h-3.5"
                  />
                </span>
              </button>
            </li>
          </ul>
        </UCard>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, watch, onMounted } from 'vue'
import AppFormField from '~/components/shared/AppFormField.vue'
import ReusableButton from '~/components/shared/ReusableButton.vue'
import Logo from '~/components/Logo.vue'
import ProvisioningJobView from '~/components/site-generator/ProvisioningJobView.vue'
import { useAppToast } from '~/composables/useAppToast'
import { useProvisioningApi, useProvisioningJob } from '~/composables/useProvisioning'
import {
  ADMIN_NAME_MAX,
  ADMIN_NAME_MIN,
  CLIENT_NAME_MAX,
  CLIENT_NAME_MIN,
  DEFAULT_BASE_DOMAIN,
  HEALTH_TIMEOUT_DEFAULT,
  HEALTH_TIMEOUT_MAX,
  HEALTH_TIMEOUT_MIN,
  PORT_MAX,
  PORT_MIN,
  buildCreateJobBody,
  buildOnboardCommand,
  classifyProvisioningError,
  isHealthTimeoutUnset,
  isTerminalJobStatus,
  isValidAdminName,
  isValidClientName,
  isValidHealthTimeout,
  isValidEmail,
  isValidHostname,
  isValidPort,
  isValidSlug,
  provisioningErrorMessage,
  suggestSlug,
  tenantDatabases,
  tenantStoragePath,
  type JobStatus,
  type NextPorts,
  type ProvisioningErrorKind,
  type ProvisioningJob,
  type ProvisioningMode
} from '~/utils/provisioning'

definePageMeta({
  layout: false,
  // Platform admin only: 02-viewer-readonly.global.ts redirects other roles (isPlatformAdminPath)
  middleware: 'auth'
})

const { success } = useAppToast()
const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const { isPlatformAdmin } = useRbac()
const api = useProvisioningApi()
const { state: jobState, attach } = useProvisioningJob()

// ---------------------------------------------------------------------------
// Wizard form (only the contract's POST /jobs fields)
// ---------------------------------------------------------------------------

const view = ref<'wizard' | 'job'>('wizard')
const currentStep = ref(1)
const showCli = ref(false)
const showAdvanced = ref(false)
const submitting = ref(false)
const submitError = ref<{ kind: ProvisioningErrorKind, message?: string } | null>(null)

const emptyForm = () => ({
  clientName: '',
  slug: '',
  mode: 'local' as ProvisioningMode,
  baseDomain: DEFAULT_BASE_DOMAIN,
  autoPorts: true,
  fePort: null as number | null,
  apiPort: null as number | null,
  adminName: '',
  adminEmail: '',
  withDemoData: false,
  provisionEmail: true,
  dryRun: false,
  allowNoTls: false,
  // null = script default (HEALTH_TIMEOUT_DEFAULT); a cleared number input yields ''
  healthTimeoutSeconds: null as number | null
})

const form = reactive(emptyForm())
const isSlugTouched = ref(false)

const steps = computed(() => [
  { id: 1, title: t('siteGenerator.steps.organization') },
  { id: 2, title: t('siteGenerator.steps.administrator') },
  { id: 3, title: t('siteGenerator.steps.infrastructure') },
  { id: 4, title: t('siteGenerator.steps.review') }
])

const modeOptions = computed(() => [
  { value: 'local' as const, icon: 'i-lucide-monitor', label: t('siteGenerator.organization.localhost') },
  { value: 'production' as const, icon: 'i-lucide-cloud', label: t('siteGenerator.organization.production') }
])

watch(() => form.clientName, (name) => {
  if (!isSlugTouched.value) form.slug = suggestSlug(name)
})

const onSlugInput = () => {
  isSlugTouched.value = true
  form.slug = form.slug.toLowerCase().replace(/[^a-z0-9]/g, '')
}

const slugOrPlaceholder = computed(() => form.slug || 'client')

const displayPorts = computed(() => form.autoPorts
  ? { fe: nextPorts.value?.fe_port, api: nextPorts.value?.api_port }
  : { fe: form.fePort ?? undefined, api: form.apiPort ?? undefined })

// Same derivation as scripts/generate-tenant-stack.sh; the job result carries the real URLs
const expectedFrontendUrl = computed(() => form.mode === 'local'
  ? `http://${slugOrPlaceholder.value}.localhost:${displayPorts.value.fe ?? '…'}`
  : `https://${slugOrPlaceholder.value}.${form.baseDomain || DEFAULT_BASE_DOMAIN}`)

const expectedApiUrl = computed(() => form.mode === 'local'
  ? `http://localhost:${displayPorts.value.api ?? '…'}`
  : `https://api-${slugOrPlaceholder.value}.${form.baseDomain || DEFAULT_BASE_DOMAIN}`)

const databases = computed(() => tenantDatabases(slugOrPlaceholder.value))
const storagePath = computed(() => tenantStoragePath(slugOrPlaceholder.value, form.mode))

const portError = (port: number | null) =>
  port === null || (port as unknown) === '' || isValidPort(port) ? undefined : t('siteGenerator.infrastructure.portError', { min: PORT_MIN, max: PORT_MAX })

const isHealthTimeoutOk = () =>
  isHealthTimeoutUnset(form.healthTimeoutSeconds) || isValidHealthTimeout(form.healthTimeoutSeconds)

const healthTimeoutError = computed(() => isHealthTimeoutOk()
  ? undefined
  : t('siteGenerator.infrastructure.healthTimeoutError', { min: HEALTH_TIMEOUT_MIN, max: HEALTH_TIMEOUT_MAX }))

const isStepValid = (step: number) => {
  if (step === 1) {
    return isValidClientName(form.clientName)
      && isValidSlug(form.slug)
      && (form.mode === 'local' || isValidHostname(form.baseDomain))
  }
  if (step === 2) return isValidAdminName(form.adminName) && isValidEmail(form.adminEmail)
  if (step === 3) {
    const portsOk = form.autoPorts || (isValidPort(form.fePort) && isValidPort(form.apiPort) && form.fePort !== form.apiPort)
    return portsOk && isHealthTimeoutOk()
  }
  return true
}

const isFormValid = computed(() => [1, 2, 3].every(isStepValid))

const reviewRows = computed(() => {
  const portsValue = form.autoPorts
    ? t('siteGenerator.review.portsAuto')
    : t('siteGenerator.review.portsValue', { frontend: form.fePort ?? '', api: form.apiPort ?? '' })
  return [
    { label: t('siteGenerator.review.organization'), value: form.clientName, mono: false },
    { label: t('siteGenerator.review.slug'), value: form.slug, mono: true },
    { label: t('siteGenerator.review.environment'), value: form.mode === 'local' ? t('siteGenerator.review.localEnvironment') : t('siteGenerator.review.productionEnvironment'), mono: false },
    { label: t('siteGenerator.review.portalUrl'), value: expectedFrontendUrl.value, mono: true },
    { label: t('siteGenerator.review.apiGateway'), value: expectedApiUrl.value, mono: true },
    { label: t('siteGenerator.review.ports'), value: portsValue, mono: true },
    { label: t('siteGenerator.review.administrator'), value: form.adminName.trim() ? `${form.adminName.trim()} (${form.adminEmail})` : form.adminEmail, mono: false },
    { label: t('siteGenerator.review.dataState'), value: form.withDemoData ? t('siteGenerator.review.demoData') : t('siteGenerator.review.emptyData'), mono: false },
    {
      label: t('siteGenerator.review.email'),
      value: form.mode === 'local' ? t('siteGenerator.review.emailSkippedLocal') : form.provisionEmail ? t('siteGenerator.review.emailYes') : t('siteGenerator.review.emailNo'),
      mono: false
    },
    ...(form.mode === 'production'
      ? [{ label: t('siteGenerator.review.tls'), value: form.allowNoTls ? t('siteGenerator.review.tlsAllowHttp') : t('siteGenerator.review.tlsRequired'), mono: false }]
      : []),
    {
      label: t('siteGenerator.review.healthTimeout'),
      value: isValidHealthTimeout(form.healthTimeoutSeconds)
        ? t('siteGenerator.review.healthTimeoutValue', { seconds: form.healthTimeoutSeconds })
        : t('siteGenerator.review.healthTimeoutDefault', { seconds: HEALTH_TIMEOUT_DEFAULT }),
      mono: false
    },
    { label: t('siteGenerator.review.storage'), value: storagePath.value, mono: true },
    { label: t('siteGenerator.review.databases'), value: databases.value.join(', '), mono: true }
  ]
})

const formInput = computed(() => ({ ...form }))
const cliCommand = computed(() => buildOnboardCommand(formInput.value))

// ---------------------------------------------------------------------------
// Ports, history, reattach
// ---------------------------------------------------------------------------

const nextPorts = ref<NextPorts | null>(null)
const portsLoading = ref(false)
const portsError = ref<ProvisioningErrorKind | null>(null)

const history = ref<ProvisioningJob[]>([])
const historyLoading = ref(false)
const historyError = ref<ProvisioningErrorKind | null>(null)

const activeJobId = ref<string | null>(null)

/** Set when the provisioner itself is down, so the page shows one banner instead of many errors. */
const serviceError = computed(() =>
  historyError.value === 'unreachable' || portsError.value === 'unreachable' || submitError.value?.kind === 'unreachable'
    ? 'unreachable'
    : null)

const loadNextPorts = async () => {
  portsLoading.value = true
  portsError.value = null
  try {
    nextPorts.value = await api.getNextPorts()
    // Prefill the manual inputs once, without overwriting what the admin typed
    if (form.fePort === null) form.fePort = nextPorts.value.fe_port
    if (form.apiPort === null) form.apiPort = nextPorts.value.api_port
  } catch (e) {
    nextPorts.value = null
    portsError.value = classifyProvisioningError(e)
  } finally {
    portsLoading.value = false
  }
}

const loadHistory = async () => {
  historyLoading.value = true
  historyError.value = null
  try {
    history.value = (await api.listJobs()).slice(0, 10)
  } catch (e) {
    historyError.value = classifyProvisioningError(e)
  } finally {
    historyLoading.value = false
  }
}

const viewJob = (job: Pick<ProvisioningJob, 'id'> & Partial<ProvisioningJob>) => {
  activeJobId.value = job.id
  view.value = 'job'
  attach(job.id, job.status ? job as ProvisioningJob : undefined)
  if (route.query.job !== job.id) router.replace({ query: { ...route.query, job: job.id } })
}

const loadInitialData = async () => {
  await Promise.all([loadHistory(), loadNextPorts()])
}

onMounted(async () => {
  if (!isPlatformAdmin.value) return
  await loadInitialData()
  // Reattach: the job in the URL first, else a job that is still queued or running
  const fromQuery = typeof route.query.job === 'string' ? route.query.job : null
  if (fromQuery) {
    viewJob(history.value.find(j => j.id === fromQuery) || { id: fromQuery })
    return
  }
  const active = history.value.find(j => !isTerminalJobStatus(j.status))
  if (active) viewJob(active)
})

// Refresh history and the next free ports when the watched job ends
watch(() => jobState.value.job?.status, (status, prev) => {
  if (isTerminalJobStatus(status) && prev && !isTerminalJobStatus(prev)) {
    void loadHistory()
    void loadNextPorts()
  }
})

// ---------------------------------------------------------------------------
// Actions
// ---------------------------------------------------------------------------

const handleProvision = async () => {
  if (!isFormValid.value || submitting.value) return
  submitting.value = true
  submitError.value = null
  try {
    const job = await api.createJob(buildCreateJobBody(formInput.value))
    viewJob(job)
    void loadHistory()
  } catch (e) {
    submitError.value = { kind: classifyProvisioningError(e), message: provisioningErrorMessage(e) }
  } finally {
    submitting.value = false
  }
}

const startNewJob = () => {
  Object.assign(form, emptyForm())
  isSlugTouched.value = false
  submitError.value = null
  showCli.value = false
  currentStep.value = 1
  view.value = 'wizard'
  activeJobId.value = null
  const { job: _job, ...rest } = route.query
  router.replace({ query: rest })
  if (nextPorts.value) {
    form.fePort = nextPorts.value.fe_port
    form.apiPort = nextPorts.value.api_port
  }
}

const copyText = async (text: string, label: string) => {
  try {
    await navigator.clipboard.writeText(text)
    success(t('siteGenerator.copy.success'), t('siteGenerator.copy.successDescription', { label }))
  } catch {
    success(t('siteGenerator.copy.fallback'), text)
  }
}

// ---------------------------------------------------------------------------
// Display helpers
// ---------------------------------------------------------------------------

const formatDate = (iso: string) => {
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleString(locale.value === 'id' ? 'id-ID' : 'en-GB')
}

const historyIcon = (s: JobStatus) => ({
  queued: 'i-lucide-hourglass',
  running: 'i-lucide-loader-2',
  succeeded: 'i-lucide-check-circle',
  failed: 'i-lucide-circle-x'
})[s] || 'i-lucide-circle'

const historyIconClass = (s: JobStatus) => ({
  queued: 'text-gray-400',
  running: 'text-primary animate-spin',
  succeeded: 'text-emerald-500',
  failed: 'text-red-500'
})[s] || ''
</script>
