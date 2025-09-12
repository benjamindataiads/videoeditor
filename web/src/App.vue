<template>
  <div class="h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 flex flex-col" :class="{ 'dark': theme === 'dark' }">
    <!-- Header -->
    <header class="border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex-shrink-0">
      <div class="flex items-center justify-between px-4 py-2">
        <h1 class="text-lg font-bold bg-gradient-to-r from-primary-600 to-accent-500 bg-clip-text text-transparent">
          Video Editor
        </h1>
        <div class="flex items-center gap-2">
          <button @click="toggleTheme" class="btn-secondary">
            <SunIcon v-if="theme === 'dark'" class="w-4 h-4" />
            <MoonIcon v-else class="w-4 h-4" />
            {{ theme === 'dark' ? 'Light' : 'Dark' }}
          </button>
          <button @click="exportTimeline" :disabled="clips.length === 0 || exporting" class="btn-primary">
            {{ exporting ? 'Exporting…' : 'Export' }}
          </button>
          <span v-if="exporting" class="text-sm text-gray-600 dark:text-gray-400">Merging clips…</span>
          <a v-if="!exporting && exportUrl" :href="backendBase + exportUrl" download
             class="btn-primary">Download</a>
        </div>
      </div>
    </header>

    <div class="flex flex-1 min-h-0">
      <!-- Sidebar -->
      <aside class="w-64 flex-shrink-0 border-r border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex flex-col">
        <div class="p-3 flex-1 overflow-y-auto">
          <div class="grid grid-cols-3 gap-1 mb-3">
            <button @click="activeTab = 'media'" 
                    :class="activeTab === 'media' ? 'btn-primary' : 'btn-secondary'">Media</button>
            <button @click="activeTab = 'audio'"
                    :class="activeTab === 'audio' ? 'btn-primary' : 'btn-secondary'">Audio</button>
            <button @click="activeTab = 'transitions'"
                    :class="activeTab === 'transitions' ? 'btn-primary' : 'btn-secondary'">FX</button>
          </div>

          <div v-if="activeTab === 'media'" class="space-y-3">
            <label class="card p-2 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
              <input type="file" accept="video/*,image/*" @change="onUpload" class="hidden" />
              <div class="flex items-center gap-2">
                <CloudArrowUpIcon class="w-3 h-3 text-primary-600" />
                <span class="text-xs">Upload Media</span>
              </div>
            </label>
            
            <div class="space-y-1">
              <div v-for="a in videoImageAssets" :key="a.id" @click="addClip(a)"
                   class="card p-2 cursor-pointer hover:shadow-md transition-all relative group">
                <img v-if="a.kind === 'image'" :src="backendBase + a.url" 
                     class="w-full h-20 object-cover rounded bg-gray-100 dark:bg-gray-700" />
                <video v-else :src="backendBase + a.url" muted
                       class="w-full h-20 object-cover rounded bg-gray-100 dark:bg-gray-700"></video>
                <p class="text-xs text-gray-600 dark:text-gray-400 mt-1 truncate">{{ a.filename }}</p>
                <button @click.stop="deleteAsset(a.id)"
                        class="absolute top-1 right-1 p-0.5 bg-black/50 text-white rounded opacity-0 group-hover:opacity-100 transition-opacity">
                  <TrashIcon class="w-3 h-3" />
                </button>
              </div>
              <div v-if="videoImageAssets.length === 0" class="text-sm text-gray-500 text-center py-8">
                No media uploaded yet
              </div>
            </div>
          </div>

          <div v-if="activeTab === 'audio'" class="space-y-3">
            <label class="card p-2 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
              <input type="file" accept="audio/*" @change="onUploadAudio" class="hidden" />
              <div class="flex items-center gap-2">
                <MusicalNoteIcon class="w-3 h-3 text-primary-600" />
                <span class="text-xs">Upload Audio</span>
              </div>
            </label>
            
            <div class="space-y-1">
              <div v-for="a in audioAssets" :key="a.id" @click="addAudioClip(a)"
                   class="card p-2 cursor-pointer hover:shadow-md transition-all relative group">
                <div class="w-full h-20 bg-gray-100 dark:bg-gray-700 rounded flex items-center justify-center">
                  <MusicalNoteIcon class="w-8 h-8 text-gray-400" />
                </div>
                <p class="text-xs text-gray-600 dark:text-gray-400 mt-1 truncate">{{ a.filename }}</p>
                <button @click.stop="deleteAsset(a.id)"
                        class="absolute top-1 right-1 p-0.5 bg-black/50 text-white rounded opacity-0 group-hover:opacity-100 transition-opacity">
                  <TrashIcon class="w-3 h-3" />
                </button>
              </div>
              <div v-if="audioAssets.length === 0" class="text-sm text-gray-500 text-center py-8">
                No audio uploaded yet
              </div>
            </div>
          </div>

          <div v-if="activeTab === 'transitions'" class="space-y-3">
            <div class="text-xs text-gray-600 dark:text-gray-400 mb-2">
              Click on a transition to add between clips
            </div>
            
            <div class="space-y-1">
              <div v-for="transition in availableTransitions" :key="transition.id" 
                   @click="selectTransition(transition)"
                   class="card p-2 cursor-pointer hover:shadow-md transition-all relative group"
                   :class="{ 'ring-2 ring-primary-500': selectedTransition?.id === transition.id }">
                <div class="w-full h-16 bg-gradient-to-r rounded flex items-center justify-center text-white text-xs font-medium"
                     :class="transition.gradient">
                  {{ transition.name }}
                </div>
                <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">{{ transition.description }}</p>
              </div>
            </div>
          </div>
        </div>
      </aside>

      <!-- Main Content -->
      <main class="flex-1 flex flex-col bg-gray-50 dark:bg-gray-900 justify-center min-w-0">
        <div class="flex flex-col items-center justify-center space-y-6 px-6">
          <!-- Aspect Ratio & Crop Controls -->
          <div class="flex items-center gap-6">
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-600 dark:text-gray-400">Aspect Ratio:</span>
              <div class="flex gap-1">
                <button @click="aspectRatio = '16:9'" 
                        :class="aspectRatio === '16:9' ? 'btn-primary' : 'btn-secondary'">16:9</button>
                <button @click="aspectRatio = '1:1'"
                        :class="aspectRatio === '1:1' ? 'btn-primary' : 'btn-secondary'">1:1</button>
                <button @click="aspectRatio = '9:16'"
                        :class="aspectRatio === '9:16' ? 'btn-primary' : 'btn-secondary'">9:16</button>
              </div>
            </div>
            
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-600 dark:text-gray-400">Fit:</span>
              <div class="flex gap-1">
                <button @click="cropMode = 'letterbox'" 
                        :class="cropMode === 'letterbox' ? 'btn-primary' : 'btn-secondary'"
                        title="Letterbox - fit content with padding">Fit</button>
                <button @click="cropMode = 'crop'"
                        :class="cropMode === 'crop' ? 'btn-primary' : 'btn-secondary'"
                        title="Crop - fill frame, may cut content">Crop</button>
              </div>
            </div>
          </div>

          <!-- Player -->
          <div :class="playerContainerClass" class="bg-gray-800 rounded-xl border border-gray-300 dark:border-gray-600 overflow-hidden shadow-lg">
            <video v-if="!isCurrentImage" ref="player" :src="currentSrc" @ended="onEnded"
                   :class="[playerMediaClass, { 'scale-x-[-1]': currentClip?.reversed }]" />
            <img v-else :class="[playerMediaClass, { 'scale-x-[-1]': currentClip?.reversed }]" :src="currentImageSrc" />
            <audio v-if="audioClips.length > 0" ref="audioPlayer" :src="audioSrc" @ended="onAudioEnded" class="hidden"/>
          </div>

          <!-- Timeline -->
          <div class="w-full">
            <div class="card p-4 min-h-[200px]">
            <div class="flex">
              <div class="w-16 flex justify-center border-r border-gray-200 dark:border-gray-700 pr-4">
                <button @click="togglePlay" 
                        class="w-12 h-12 rounded-full bg-primary-600 text-white hover:bg-primary-700 transition-colors flex items-center justify-center">
                  <PlayIcon v-if="!isPlaying" class="w-6 h-6 ml-0.5" />
                  <StopIcon v-else class="w-6 h-6" />
                </button>
              </div>
              
              <div class="flex-1 relative pl-4 overflow-x-auto" ref="timelineEl" @mousemove="onTimelineMove" @mouseleave="hideCursor">
                <!-- Ruler -->
                <div class="absolute top-2 left-4 h-4 pointer-events-none" :style="{ width: (trackLeftPad + totalWidth) + 'px' }">
                  <div v-for="t in ticks" :key="t" class="absolute top-0 text-xs text-gray-500"
                       :style="{ left: (trackLeftPad + t * pxPerSec) + 'px', transform: 'translateX(-50%)' }">
                    {{ formatTime(t) }}
                  </div>
                </div>

                <!-- Video Clips -->
                <div class="pt-6 flex gap-2" :style="{ width: totalWidth + 'px', marginLeft: trackLeftPad + 'px' }">
                  <template v-for="(c, i) in clips" :key="i">
                    <div @click="playAt(i)"
                         class="timeline-clip group relative"
                         :class="{ 'active': i === currentIndex }"
                         :style="{ width: clipWidth(c) + 'px' }">
                      <img v-if="getAsset(c.assetId)?.kind === 'image'" 
                           :src="backendBase + (getAsset(c.assetId)?.url || '')"
                           class="w-20 h-14 object-cover rounded bg-gray-100"
                           :class="{ 'scale-x-[-1]': c.reversed }" />
                      <video v-else :src="backendBase + (getAsset(c.assetId)?.url || '')" muted preload="metadata"
                             class="w-20 h-14 object-cover rounded bg-gray-100"
                             :class="{ 'scale-x-[-1]': c.reversed }"></video>
                      
                      <div class="absolute top-0 left-0 w-1.5 h-full bg-gray-300 dark:bg-gray-600 cursor-ew-resize rounded-l"
                           @mousedown.stop.prevent="beginTrim(i, 'left', $event)"></div>
                      <div class="absolute top-0 right-0 w-1.5 h-full bg-gray-300 dark:bg-gray-600 cursor-ew-resize rounded-r"
                           @mousedown.stop.prevent="beginTrim(i, 'right', $event)"></div>
                      
                      <!-- Mirror Toggle Button -->
                      <button @click.stop="toggleMirror(i)"
                              class="absolute top-1 left-1 w-5 h-5 bg-blue-500 text-white rounded text-xs hover:bg-blue-600 transition-all opacity-0 group-hover:opacity-100"
                              :class="{ 'opacity-100 bg-blue-600': c.reversed }"
                              title="Toggle horizontal mirror">
                        ⇄
                      </button>
                      
                      <!-- Reverse Playback Toggle Button (only for videos) -->
                      <button v-if="getAsset(c.assetId)?.kind === 'video'" 
                              @click.stop="toggleReversePlayback(i)"
                              class="absolute top-1 left-7 w-5 h-5 bg-purple-500 text-white rounded text-xs hover:bg-purple-600 transition-all opacity-0 group-hover:opacity-100"
                              :class="{ 'opacity-100 bg-purple-600': c.reversePlayback }"
                              title="Toggle reverse playback">
                        ⏪
                      </button>
                      
                      <span class="text-xs text-gray-600 dark:text-gray-400 mt-1">{{ displayDuration(c).toFixed(1) }}s</span>
                      <button @click.stop="removeClip(i)"
                              class="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white rounded-full text-xs hover:bg-red-600 transition-colors">
                        ✕
                      </button>
                    </div>

                    <!-- Transition between clips -->
                    <div v-if="i < clips.length - 1" class="transition-area relative flex items-center">
                      <!-- Add Transition Button -->
                      <button v-if="!getTransitionForClip(i) && selectedTransition"
                              @click="addTransitionBetweenClips(i)"
                              class="w-6 h-6 bg-orange-500 text-white rounded-full text-xs hover:bg-orange-600 transition-colors flex items-center justify-center"
                              :title="`Add ${selectedTransition.name} transition`">
                        +
                      </button>
                      
                      <!-- Debug info -->
                      <div v-if="getTransitionForClip(i)" class="text-xs text-gray-500 mt-1">
                        {{ getTransitionForClip(i)?.name }}
                      </div>
                      
                      <!-- Existing Transition -->
                      <div v-else-if="getTransitionForClip(i)" 
                           class="transition-indicator relative group">
                        <div class="w-8 h-4 rounded flex items-center justify-center text-xs text-white font-medium"
                             :class="getTransitionForClip(i)?.gradient"
                             :title="getTransitionForClip(i)?.name">
                          FX
                        </div>
                        <button @click="removeTransition(i)"
                                class="absolute -top-1 -right-1 w-3 h-3 bg-red-500 text-white rounded-full text-xs hover:bg-red-600 transition-colors opacity-0 group-hover:opacity-100">
                          ✕
                        </button>
                      </div>
                      
                      <!-- Placeholder when no transition selected -->
                      <div v-else class="w-6 h-6 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-full flex items-center justify-center">
                        <span class="text-xs text-gray-400">+</span>
                      </div>
                    </div>
                  </template>
                  
                  <div v-if="clips.length === 0" class="text-sm text-gray-500 py-8">
                    Add clips from the sidebar
                  </div>
                </div>

                <!-- Audio Track -->
                <div v-if="audioClips.length > 0" class="mt-2 flex gap-2" :style="{ width: totalWidth + 'px', marginLeft: trackLeftPad + 'px' }">
                  <div v-for="(ac, i) in audioClips" :key="i"
                       class="bg-green-100 dark:bg-green-900 border border-green-200 dark:border-green-700 rounded-lg px-3 py-2 flex items-center gap-2 h-8"
                       :style="{ width: audioClipWidth(ac) + 'px' }">
                    <MusicalNoteIcon class="w-4 h-4 text-green-600 dark:text-green-400" />
                    <span class="text-xs text-green-700 dark:text-green-300 truncate max-w-24">
                      {{ getAsset(ac.assetId)?.filename || 'Audio' }}
                    </span>
                    <button @click.stop="removeAudioClip(i)"
                            class="w-4 h-4 bg-red-500 text-white rounded-full text-xs hover:bg-red-600 transition-colors flex items-center justify-center">
                      ✕
                    </button>
                  </div>
                </div>

                <!-- Cursor and Playhead -->
                <div v-show="cursor.visible" class="absolute top-0 bottom-0 w-0.5 bg-primary-500 pointer-events-none"
                     :style="{ left: cursor.x + 'px' }">
                  <div class="absolute -top-5 -left-8 bg-gray-800 text-white text-xs px-2 py-1 rounded">
                    {{ formatTime(cursor.timeSec) }}
                  </div>
                </div>
                <div class="absolute top-0 bottom-0 w-0.5 bg-accent-500 pointer-events-none" :style="{ left: playheadX + 'px' }"></div>
              </div>
            </div>
          </div>
        </div>
        </div>
      </main>

      <!-- Exports Sidebar -->
      <aside class="w-72 flex-shrink-0 border-l border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex flex-col">
        <div class="p-4 flex-1 overflow-y-auto">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Exports</h3>
            <button @click="loadExports" class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded">
              <ArrowPathIcon class="w-4 h-4" />
            </button>
          </div>
          
          <div class="space-y-2">
            <div v-for="e in exports" :key="e.url" class="card p-3">
              <div class="flex items-center justify-between">
                <span class="text-xs text-gray-600 dark:text-gray-400 truncate flex-1" :title="e.filename">
                  {{ e.filename }}
                </span>
                <div class="flex gap-1 ml-2">
                  <a :href="backendBase + e.url" target="_blank" rel="noopener noreferrer"
                     class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded">
                    <PlayIcon class="w-4 h-4" />
                  </a>
                  <a :href="backendBase + e.url" download target="_blank" rel="noopener noreferrer"
                     class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded">
                    <ArrowDownTrayIcon class="w-4 h-4" />
                  </a>
                  <button @click="deleteExport(e)"
                          class="p-1.5 text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded">
                    <TrashIcon class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
            <div v-if="exports.length === 0" class="text-sm text-gray-500 text-center py-8">
              No exports yet
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import axios from 'axios'
import { computed, nextTick, onMounted, onBeforeUnmount, ref } from 'vue'
import {
  SunIcon, MoonIcon, PlayIcon, StopIcon, CloudArrowUpIcon, MusicalNoteIcon,
  TrashIcon, ArrowPathIcon, ArrowDownTrayIcon
} from '@heroicons/vue/24/outline'

type Asset = { id: string; filename: string; url: string; kind: 'image'|'video'|'audio'|'unknown' }
type ExportClip = { assetId: string; startSec?: number; endSec?: number; durationSec?: number; reversed?: boolean; reversePlayback?: boolean }
type ExportItem = { filename: string; url: string; size?: number; modTime?: string }
type Transition = { id: string; name: string; description: string; gradient: string; ffmpegFilter: string; duration: number }
type ClipTransition = { transitionId: string; duration: number }

// Debug environment variables
console.log('Environment variables:', import.meta.env)
console.log('VITE_BACKEND_BASE:', import.meta.env.VITE_BACKEND_BASE)

const backendBase = import.meta.env.VITE_BACKEND_BASE || 'https://videoeditor-production-3bd0.up.railway.app'
console.log('Using backend URL:', backendBase)

const assets = ref<Asset[]>([])
const assetDurations = ref<Record<string, number>>({})
const clips = ref<ExportClip[]>([])
const currentIndex = ref(0)
const player = ref<HTMLVideoElement | null>(null)
const audioPlayer = ref<HTMLAudioElement | null>(null)
const isPlaying = ref(false)
const exporting = ref(false)
const exportUrl = ref('')
const timelineEl = ref<HTMLDivElement | null>(null)
const theme = ref<'dark'|'light'>((localStorage.getItem('ve-theme') as 'dark'|'light') || 'dark')
const exports = ref<ExportItem[]>([])
const playheadX = ref(0)
const activeTab = ref<'media'|'audio'|'transitions'>('media')
const aspectRatio = ref<'16:9'|'1:1'|'9:16'>('16:9')
const cropMode = ref<'letterbox'|'crop'>('letterbox')
let animId: number | null = null
let imageStartMs = 0
let imageDurSec = 1

const currentClip = computed(() => clips.value[currentIndex.value])
const videoImageAssets = computed(() => assets.value.filter(a => a.kind === 'video' || a.kind === 'image'))
const audioAssets = computed(() => assets.value.filter(a => a.kind === 'audio'))
const audioClips = ref<ExportClip[]>([])
const currentAsset = computed(() => currentClip.value ? assets.value.find(x => x.id === currentClip.value!.assetId) : undefined)
const isCurrentImage = computed(() => currentAsset.value?.kind === 'image')
const currentSrc = computed(() => isCurrentImage.value ? '' : (currentAsset.value ? backendBase + currentAsset.value.url : ''))
const currentImageSrc = computed(() => isCurrentImage.value && currentAsset.value ? backendBase + currentAsset.value.url : '')
const audioSrc = computed(() => audioClips.value.length > 0 ? backendBase + getAsset(audioClips.value[0].assetId)?.url : '')

// Transitions
const selectedTransition = ref<Transition | null>(null)
const clipTransitions = ref<Record<number, ClipTransition>>({}) // clipIndex -> transition
const availableTransitions = ref<Transition[]>([
  {
    id: 'fade',
    name: 'Fade',
    description: 'Smooth crossfade transition',
    gradient: 'from-black to-gray-600',
    ffmpegFilter: 'fade',
    duration: 1.0
  },
  {
    id: 'dissolve',
    name: 'Dissolve',
    description: 'Dissolve between clips',
    gradient: 'from-blue-500 to-blue-700',
    ffmpegFilter: 'dissolve',
    duration: 1.5
  },
  {
    id: 'wipeleft',
    name: 'Wipe Left',
    description: 'Wipe from right to left',
    gradient: 'from-green-500 to-green-700',
    ffmpegFilter: 'wipeleft',
    duration: 1.0
  },
  {
    id: 'wiperight',
    name: 'Wipe Right',
    description: 'Wipe from left to right',
    gradient: 'from-green-700 to-green-500',
    ffmpegFilter: 'wiperight',
    duration: 1.0
  },
  {
    id: 'slideleft',
    name: 'Slide Left',
    description: 'Slide from right to left',
    gradient: 'from-purple-500 to-purple-700',
    ffmpegFilter: 'slideleft',
    duration: 1.2
  },
  {
    id: 'slideright',
    name: 'Slide Right',
    description: 'Slide from left to right',
    gradient: 'from-purple-700 to-purple-500',
    ffmpegFilter: 'slideright',
    duration: 1.2
  },
  {
    id: 'circlecrop',
    name: 'Circle',
    description: 'Circular crop transition',
    gradient: 'from-orange-500 to-orange-700',
    ffmpegFilter: 'circlecrop',
    duration: 1.5
  },
  {
    id: 'radial',
    name: 'Radial',
    description: 'Radial wipe transition',
    gradient: 'from-red-500 to-red-700',
    ffmpegFilter: 'radial',
    duration: 1.3
  }
])

const playerContainerClass = computed(() => {
  switch (aspectRatio.value) {
    case '16:9': return 'w-[640px] h-[360px]'
    case '1:1': return 'w-[400px] h-[400px]'
    case '9:16': return 'w-[360px] h-[640px]'
    default: return 'w-[640px] h-[360px]'
  }
})

const playerMediaClass = computed(() => {
  const base = 'w-full h-full bg-gray-800'
  if (cropMode.value === 'crop') {
    return `${base} object-cover`
  } else {
    return `${base} object-contain`
  }
})

function getAsset(id: string) { return assets.value.find(a => a.id === id) }

async function fetchAssets() {
  const { data } = await axios.get<Asset[]>(backendBase + '/api/assets')
  assets.value = data
  for (const a of assets.value) {
    if (a.kind === 'video') probeDuration(a)
    if (a.kind === 'audio') probeAudioDuration(a)
  }
}

async function onUpload(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files || input.files.length === 0) return
  const file = input.files[0]
  const form = new FormData()
  form.append('file', file)
  const { data } = await axios.post<Asset>(backendBase + '/api/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
  assets.value.unshift(data)
  if (data.kind === 'video') probeDuration(data)
  input.value = ''
}

async function onUploadAudio(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files || input.files.length === 0) return
  const file = input.files[0]
  const form = new FormData()
  form.append('file', file)
  const { data } = await axios.post<Asset>(backendBase + '/api/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
  assets.value.unshift(data)
  if (data.kind === 'audio') {
    activeTab.value = 'audio'
    probeAudioDuration(data)
  }
  input.value = ''
}

function addAudioClip(a: Asset) {
  const audioDur = assetDurations.value[a.id] || totalSeconds.value || 30
  audioClips.value.push({ assetId: a.id, durationSec: audioDur })
}

function removeAudioClip(i: number) {
  audioClips.value.splice(i, 1)
}

function audioClipWidth(ac: ExportClip) {
  return (ac.durationSec ?? totalSeconds.value) * pxPerSec
}

async function deleteAsset(id: string) {
  await axios.delete(backendBase + '/api/assets/' + id)
  assets.value = assets.value.filter(a => a.id !== id)
}

function addClip(a: Asset) {
  if (a.kind === 'image') {
    clips.value.push({ assetId: a.id, durationSec: 1 })
  } else {
    const d = assetDurations.value[a.id] ?? 0
    clips.value.push({ assetId: a.id, startSec: 0, endSec: d, durationSec: d })
  }
  if (clips.value.length === 1) playAt(0)
}

function removeClip(i: number) {
  clips.value.splice(i, 1)
  if (currentIndex.value >= clips.value.length) currentIndex.value = clips.value.length - 1
  playAt(Math.max(0, currentIndex.value))
}

function toggleMirror(i: number) {
  const clip = clips.value[i]
  clip.reversed = !clip.reversed
  // Refresh the current player if this is the active clip
  if (i === currentIndex.value) {
    playAt(i)
  }
}

function toggleReversePlayback(i: number) {
  const clip = clips.value[i]
  clip.reversePlayback = !clip.reversePlayback
  // Refresh the current player if this is the active clip
  if (i === currentIndex.value) {
    playAt(i)
  }
}

// Transition functions
function selectTransition(transition: Transition) {
  selectedTransition.value = transition
}

function addTransitionBetweenClips(clipIndex: number) {
  if (!selectedTransition.value || clipIndex >= clips.value.length - 1) return
  
  clipTransitions.value[clipIndex] = {
    transitionId: selectedTransition.value.id,
    duration: selectedTransition.value.duration
  }
  
  console.log('DEBUG: Added transition', selectedTransition.value.name, 'between clips', clipIndex, 'and', clipIndex + 1)
}

function removeTransition(clipIndex: number) {
  delete clipTransitions.value[clipIndex]
}

function getTransitionForClip(clipIndex: number): Transition | null {
  const clipTransition = clipTransitions.value[clipIndex]
  if (!clipTransition) return null
  return availableTransitions.value.find(t => t.id === clipTransition.transitionId) || null
}

async function playAt(i: number) {
  currentIndex.value = i
  await nextTick()
  clearImageTimer()
  detachVideoGuards()
  clearReversePlayback()
  
  // Start audio if present
  if (audioPlayer.value && audioClips.value.length > 0) {
    const timelineStart = getTimelineStartForClip(i)
    audioPlayer.value.currentTime = timelineStart
    audioPlayer.value.play()
  }
  
  if (isCurrentImage.value) { imageStartMs = performance.now(); imageDurSec = currentClip.value?.durationSec ?? 1; startImageTimer(); startAnim() }
  else if (player.value) {
    const clip = currentClip.value!
    const start = clip.startSec ?? 0
    const end = (clip.endSec ?? (clip.durationSec ?? 0))
    const p = player.value
    
    if (clip.reversePlayback) {
      // Handle reverse playback
      const ensureReverseStart = () => {
        p.currentTime = Math.min(end, p.duration || end)
        startReversePlayback(p, start, end)
      }
      
      if (isNaN(p.duration) || !isFinite(p.duration)) {
        const onMeta = () => { p.removeEventListener('loadedmetadata', onMeta); ensureReverseStart() }
        p.addEventListener('loadedmetadata', onMeta)
        p.load()
      } else {
        ensureReverseStart()
      }
    } else {
      // Normal forward playback
      attachVideoGuards(end)
      const ensureStart = () => {
        p.currentTime = Math.max(0, start)
        p.play()
      }
      // track playhead
      p.ontimeupdate = () => updatePlayhead()
      startAnim()
      if (isNaN(p.duration) || !isFinite(p.duration)) {
        const onMeta = () => { p.removeEventListener('loadedmetadata', onMeta); ensureStart() }
        p.addEventListener('loadedmetadata', onMeta)
        p.load()
      } else {
        ensureStart()
      }
    }
  }
}

async function loadExports() {
  const { data } = await axios.get<ExportItem[]>(backendBase + '/api/exports')
  exports.value = data
}

async function deleteExport(e: ExportItem) {
  const filename = e.filename || e.url.split('/').pop()!
  await axios.delete(backendBase + '/api/exports/' + encodeURIComponent(filename))
  await loadExports()
}

function onEnded() {
  const next = currentIndex.value + 1
  if (next < clips.value.length) playAt(next)
  else {
    isPlaying.value = false
    stopAnim()
    // Stop audio when timeline ends
    if (audioPlayer.value) audioPlayer.value.pause()
  }
}

async function exportTimeline() {
  exporting.value = true
  exportUrl.value = ''
  
  // Debug: log transitions being sent
  const transitions = Object.entries(clipTransitions.value).map(([clipIndex, transition]) => ({
    clipIndex: parseInt(clipIndex),
    transitionId: transition.transitionId,
    duration: transition.duration
  }))
  console.log('DEBUG: Sending transitions:', transitions)
  
  try {
    const { data } = await axios.post<{exportId:string; url:string; status:string; error?:string}>(backendBase + '/api/export', {
      clips: clips.value.map(c => ({
        assetId: c.assetId,
        startSec: c.startSec ?? 0,
        endSec: c.endSec ?? 0,
        durationSec: c.durationSec ?? 0,
        reversed: c.reversed ?? false,
        reversePlayback: c.reversePlayback ?? false
      })),
      audio: audioClips.value.length > 0 ? { assetId: audioClips.value[0].assetId, volume: 1 } : undefined,
      aspectRatio: aspectRatio.value,
      cropMode: cropMode.value,
      transitions: Object.entries(clipTransitions.value).map(([clipIndex, transition]) => ({
        clipIndex: parseInt(clipIndex),
        transitionId: transition.transitionId,
        duration: transition.duration
      }))
    })
    if (data.status === 'done' && data.url) {
      exportUrl.value = data.url
      return
    }
    const id = data.exportId
    for (;;) {
      await new Promise(r => setTimeout(r, 800))
      const res = await axios.get<{exportId:string; url:string; status:string; error?:string}>(backendBase + '/api/export/' + id)
      if (res.data.status === 'done') { exportUrl.value = res.data.url; break }
      if (res.data.status === 'error') { throw new Error(res.data.error || 'Export failed') }
    }
  } finally {
    exporting.value = false
  }
}

onMounted(() => { fetchAssets(); applyTheme(); loadExports() })
onBeforeUnmount(() => { endDrag(); clearImageTimer() })

// Image playback handling
let imageTimer: number | undefined
function startImageTimer() {
  const dur = (currentClip.value?.durationSec ?? 1) * 1000
  imageTimer = window.setTimeout(() => onEnded(), dur)
}
function clearImageTimer() {
  if (imageTimer) window.clearTimeout(imageTimer)
  imageTimer = undefined
}

// Video end guards for trimmed playback
let guardEnd = 0
function onTimeUpdateGuard(this: HTMLVideoElement) {
  if (this.currentTime >= guardEnd - 0.05) {
    this.pause()
    this.removeEventListener('timeupdate', onTimeUpdateGuard)
    onEnded()
  }
}
function attachVideoGuards(end: number) {
  guardEnd = end
  if (player.value) player.value.addEventListener('timeupdate', onTimeUpdateGuard)
}
function detachVideoGuards() {
  if (player.value) player.value.removeEventListener('timeupdate', onTimeUpdateGuard)
}

// Reverse playback implementation
let reverseInterval: number | null = null
let reversePlaybackStart: number = 0
let reversePlaybackDuration: number = 0

function startReversePlayback(video: HTMLVideoElement, startTime: number, endTime: number) {
  video.pause() // Don't use native playback
  const fps = 30 // Target 30 FPS for smooth reverse playback
  const frameTime = 1000 / fps
  const playbackRate = 1 / fps // Move backwards by 1/30th of a second per frame
  
  let currentTime = video.currentTime
  reversePlaybackStart = performance.now()
  reversePlaybackDuration = endTime - startTime
  
  reverseInterval = window.setInterval(() => {
    currentTime -= playbackRate
    
    if (currentTime <= startTime) {
      // Reached the beginning, move to next clip
      clearReversePlayback()
      onEnded()
      return
    }
    
    video.currentTime = currentTime
    updatePlayhead()
  }, frameTime)
  
  startAnim()
}

function clearReversePlayback() {
  if (reverseInterval) {
    window.clearInterval(reverseInterval)
    reverseInterval = null
  }
  reversePlaybackStart = 0
  reversePlaybackDuration = 0
}

// Custom controls
function playFromCurrent() { playAt(currentIndex.value) }
function togglePlay() {
  if (!isPlaying.value) {
    playAt(0)
    isPlaying.value = true
    startAnim()
  } else {
    // Stop playback without triggering playAt
    if (player.value) player.value.pause()
    if (audioPlayer.value) audioPlayer.value.pause()
    clearImageTimer()
    detachVideoGuards()
    clearReversePlayback()
    currentIndex.value = 0
    if (player.value) player.value.currentTime = 0
    if (audioPlayer.value) audioPlayer.value.currentTime = 0
    isPlaying.value = false
    stopAnim()
    playheadX.value = trackLeftPad
  }
}
function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  localStorage.setItem('ve-theme', theme.value)
  applyTheme()
}
function applyTheme() { 
  document.documentElement.classList.toggle('dark', theme.value === 'dark')
}

// Timeline sizing and ruler
const pxPerSec = 80
const trackLeftPad = 20
function displayDuration(c: ExportClip) {
  const a = getAsset(c.assetId)
  if (a?.kind === 'video') {
    const start = c.startSec ?? 0
    const end = c.endSec ?? assetDurations.value[c.assetId] ?? 1
    return Math.max(0.1, end - start)
  }
  if (c.durationSec && c.durationSec > 0) return c.durationSec
  return 1
}
function clipWidth(c: ExportClip) { return displayDuration(c) * pxPerSec }
const totalWidth = computed(() => clips.value.reduce((w, c) => w + clipWidth(c), 0))
const totalSeconds = computed(() => clips.value.reduce((s, c) => s + displayDuration(c), 0))
const ticks = computed(() => {
  const totalSec = Math.ceil(clips.value.reduce((s, c) => s + displayDuration(c), 0))
  return Array.from({ length: totalSec + 1 }, (_, i) => i)
})
function formatTime(t: number) { return `${Math.floor(t/60)}:${String(Math.floor(t%60)).padStart(2,'0')}` }

// Hover cursor preview
const cursor = ref<{ visible: boolean; x: number; timeSec: number }>({ visible: false, x: 0, timeSec: 0 })
function onTimelineMove(e: MouseEvent) {
  if (!timelineEl.value) return
  const rect = timelineEl.value.getBoundingClientRect()
  const x = e.clientX - rect.left
  cursor.value.visible = true
  cursor.value.x = x
  const time = x / pxPerSec
  cursor.value.timeSec = Math.max(0, time)
}
function hideCursor() { cursor.value.visible = false }

function updatePlayhead() {
  // elapsed before current clip
  let elapsed = 0
  for (let i = 0; i < currentIndex.value; i++) elapsed += displayDuration(clips.value[i])
  let within = 0
  const p = player.value
  const currentClipData = currentClip.value
  
  if (isCurrentImage.value) {
    if (imageStartMs > 0) {
      within = Math.min((performance.now() - imageStartMs) / 1000, imageDurSec)
    }
  } else if (p && !isNaN(p.currentTime) && currentClipData) {
    if (currentClipData.reversePlayback && reverseInterval) {
      // For reverse playback, calculate timeline position based on elapsed time
      // The playhead should move left to right even though video plays backwards
      const elapsedSeconds = (performance.now() - reversePlaybackStart) / 1000
      within = Math.min(elapsedSeconds, reversePlaybackDuration)
    } else {
      // Normal forward playback
      const start = currentClipData.startSec ?? 0
      within = Math.max(0, p.currentTime - start)
    }
  }
  playheadX.value = trackLeftPad + (elapsed + within) * pxPerSec
}

function startAnim() {
  stopAnim()
  const tick = () => { updatePlayhead(); animId = requestAnimationFrame(tick) }
  animId = requestAnimationFrame(tick)
}
function stopAnim() {
  if (animId) cancelAnimationFrame(animId)
  animId = null
}

// Trim drag state
let drag: null | { index: number; side: 'left'|'right'; startX: number; startDur: number; startStart?: number; startEnd?: number; kind: 'image'|'video'|'audio'|'unknown' } = null
function beginTrim(i: number, side: 'left'|'right', ev: MouseEvent) {
  const c = clips.value[i]
  const a = getAsset(c.assetId)
  drag = { index: i, side, startX: ev.clientX, startDur: displayDuration(c), startStart: c.startSec ?? 0, startEnd: c.endSec ?? (assetDurations.value[c.assetId] ?? (c.durationSec ?? 1)), kind: a?.kind ?? 'unknown' }
  window.addEventListener('mousemove', onDrag)
  window.addEventListener('mouseup', endDrag)
}
function onDrag(ev: MouseEvent) {
  if (!drag) return
  const dx = (ev.clientX - drag.startX) / pxPerSec
  const c = clips.value[drag.index]
  if (drag.kind === 'image') {
    let newDur = Math.max(0.2, drag.startDur + (drag.side === 'right' ? dx : -dx))
    c.durationSec = newDur
  } else {
    const maxDur = assetDurations.value[c.assetId] ?? drag.startEnd ?? (drag.startStart ?? 0)
    if (drag.side === 'left') {
      const newStart = Math.max(0, Math.min(drag.startStart! + dx, (drag.startEnd! - 0.2)))
      c.startSec = newStart
    } else {
      const newEnd = Math.min(maxDur, Math.max(drag.startEnd! + dx, (drag.startStart! + 0.2)))
      c.endSec = newEnd
    }
    c.durationSec = Math.max(0.2, (c.endSec ?? drag.startEnd!) - (c.startSec ?? drag.startStart!))
  }
}
function endDrag() {
  if (!drag) return
  window.removeEventListener('mousemove', onDrag)
  window.removeEventListener('mouseup', endDrag)
  drag = null
}

// Duration probing via <video> metadata (frontend-only fallback)
function probeDuration(a: Asset) {
  const v = document.createElement('video')
  v.preload = 'metadata'
  v.src = backendBase + a.url
  v.muted = true
  v.addEventListener('loadedmetadata', () => {
    if (!isFinite(v.duration)) return
    assetDurations.value[a.id] = v.duration
  }, { once: true })
}

function probeAudioDuration(a: Asset) {
  const audio = document.createElement('audio')
  audio.preload = 'metadata'
  audio.src = backendBase + a.url
  audio.addEventListener('loadedmetadata', () => {
    if (!isFinite(audio.duration)) return
    assetDurations.value[a.id] = audio.duration
  }, { once: true })
}

function onAudioEnded() {
  // Audio ended, but continue with video timeline
}

function getTimelineStartForClip(clipIndex: number) {
  let elapsed = 0
  for (let i = 0; i < clipIndex; i++) {
    elapsed += displayDuration(clips.value[i])
  }
  return elapsed
}
</script>