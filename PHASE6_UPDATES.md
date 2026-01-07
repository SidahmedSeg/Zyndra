# Phase 6 Updates - Technology Stack Changes

## Summary

Based on feedback, Phase 6 has been updated with the following changes:

## Changes Made

### 1. **Next.js Instead of Plain React**
- ✅ Changed from React + Vite to **Next.js 14+** with App Router
- ✅ Better SEO, server-side rendering, and built-in routing
- ✅ Better performance and developer experience

### 2. **Bun Instead of Node.js**
- ✅ Use **Bun** as package manager and runtime
- ✅ Faster package installation and execution
- ✅ Native TypeScript support
- ✅ Better performance

### 3. **Large Drawers Instead of Modals**
- ✅ Changed from modal dialogs to **large side panel drawers**
- ✅ Drawer width: ~800px (much larger than typical modals)
- ✅ Better UX for complex configuration forms
- ✅ More space for tabs and content
- ✅ Use shadcn/ui Drawer component or custom implementation

## Updated Technology Stack

### Frontend Stack
- **Framework:** Next.js 14+ (App Router)
- **Runtime:** Bun
- **UI Components:** 
  - React Flow (canvas)
  - Tailwind CSS (styling)
  - shadcn/ui (drawer components)
  - Lucide React (icons)
- **State Management:** Zustand
- **HTTP Client:** Axios or native fetch
- **Real-time:** Centrifugo

## Updated Phase 6 Tasks

### 6.1 Frontend Setup
- Initialize Next.js project using Bun
- Set up Next.js 14+ with App Router
- Configure Bun as package manager and runtime
- Install dependencies (see above)

### 6.6 Configuration Drawers (Updated)
- Create large drawer components (side panels):
  - `ServiceDrawer.tsx` (~800px width)
  - `DatabaseDrawer.tsx` (~800px width)
  - `VolumeDrawer.tsx` (~800px width)
- Use shadcn/ui Drawer component
- Drawer slides in from the right side
- Full-height drawers for better UX

## Benefits

1. **Next.js:**
   - Better performance with SSR/SSG
   - Built-in routing
   - Better SEO
   - Server components support

2. **Bun:**
   - Faster package installation
   - Faster runtime execution
   - Native TypeScript support
   - Better compatibility with Node.js ecosystem

3. **Large Drawers:**
   - More space for configuration
   - Better UX for complex forms
   - Less claustrophobic than modals
   - Better for multi-tab interfaces

## Implementation Notes

### Next.js Setup
```bash
# Using Bun
bun create next-app web
cd web
bun install
```

### Drawer Implementation
- Use shadcn/ui Drawer component
- Width: ~800px (configurable)
- Slide in from right
- Full height
- Backdrop overlay
- Close on escape key

### Bun Configuration
- Use `bun` instead of `npm` or `yarn`
- Bun works with existing Node.js packages
- Faster installs and execution

## Updated Files

- `DEVELOPMENT_PLAN.md` - Phase 6 updated
- `PROJECT_STATUS.md` - Technology stack updated

