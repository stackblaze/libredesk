import { cva } from 'class-variance-authority';

export { default as Sidebar } from './Sidebar.vue';
export { default as SidebarContent } from './SidebarContent.vue';
export { default as SidebarFooter } from './SidebarFooter.vue';
export { default as SidebarGroup } from './SidebarGroup.vue';
export { default as SidebarGroupAction } from './SidebarGroupAction.vue';
export { default as SidebarGroupContent } from './SidebarGroupContent.vue';
export { default as SidebarGroupLabel } from './SidebarGroupLabel.vue';
export { default as SidebarHeader } from './SidebarHeader.vue';
export { default as SidebarInput } from './SidebarInput.vue';
export { default as SidebarInset } from './SidebarInset.vue';
export { default as SidebarMenu } from './SidebarMenu.vue';
export { default as SidebarMenuAction } from './SidebarMenuAction.vue';
export { default as SidebarMenuBadge } from './SidebarMenuBadge.vue';
export { default as SidebarMenuButton } from './SidebarMenuButton.vue';
export { default as SidebarMenuItem } from './SidebarMenuItem.vue';
export { default as SidebarMenuSkeleton } from './SidebarMenuSkeleton.vue';
export { default as SidebarMenuSub } from './SidebarMenuSub.vue';
export { default as SidebarMenuSubButton } from './SidebarMenuSubButton.vue';
export { default as SidebarMenuSubItem } from './SidebarMenuSubItem.vue';
export { default as SidebarProvider } from './SidebarProvider.vue';
export { default as SidebarRail } from './SidebarRail.vue';
export { default as SidebarSeparator } from './SidebarSeparator.vue';
export { default as SidebarTrigger } from './SidebarTrigger.vue';

export { useSidebar } from './utils';

export const sidebarMenuButtonVariants = cva(
  // OpenBooks rail language: quiet muted rows, green fill only when active,
  // soft rounded-lg pills, compact 13.5px type.
  'peer/menu-button flex w-full items-center gap-2.5 overflow-hidden rounded-lg px-2.5 py-[7px] text-left text-[13.5px] font-medium outline-none ring-sidebar-ring transition-colors text-foreground/70 hover:bg-muted hover:text-foreground focus-visible:ring-2 active:bg-muted active:text-foreground disabled:pointer-events-none disabled:opacity-50 group-has-[[data-sidebar=menu-action]]/menu-item:pr-8 aria-disabled:pointer-events-none aria-disabled:opacity-50 data-[active=true]:bg-[hsl(var(--ob-green-50))] data-[active=true]:font-semibold data-[active=true]:text-[hsl(var(--ob-green-800))] data-[state=open]:hover:bg-muted data-[state=open]:hover:text-foreground group-data-[collapsible=icon]:!size-8 group-data-[collapsible=icon]:!p-2 [&>span:last-child]:truncate [&>svg]:size-[17px] [&>svg]:shrink-0 [&>svg]:opacity-85',
  {
    variants: {
      variant: {
        default: 'hover:bg-muted hover:text-foreground',
        outline:
          'bg-background shadow-[0_0_0_1px_hsl(var(--sidebar-border))] hover:bg-muted hover:text-foreground hover:shadow-[0_0_0_1px_hsl(var(--sidebar-border))]',
      },
      size: {
        default: 'h-auto min-h-8 max-md:min-h-11 text-[13.5px]',
        // Nested leaves still use the OpenBooks green active pill (not muted gray).
        sm: 'h-auto min-h-7 rounded-md px-2 py-1.5 text-[12.5px] font-medium data-[active=true]:bg-[hsl(var(--ob-green-50))] data-[active=true]:font-semibold data-[active=true]:text-[hsl(var(--ob-green-800))]',
        lg: 'h-12 text-sm group-data-[collapsible=icon]:!p-0',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
);
