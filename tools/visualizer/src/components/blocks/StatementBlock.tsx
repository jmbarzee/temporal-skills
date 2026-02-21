import type { Statement } from '../../types/ast'
import { ActivityCallBlock, WorkflowCallBlock } from './CallBlocks'
import { AwaitStmtBlock, AwaitAllBlockComponent, AwaitOneBlockComponent } from './AwaitBlocks'
import { SwitchBlockComponent, IfBlock, ForBlock } from './ControlFlowBlocks'
import { ReturnBlock, CloseBlock, RawBlock, SimpleBlock, PromiseBlock, SetBlock, UnsetBlock } from './LeafBlocks'
import './blocks.css'

interface StatementBlockProps {
  statement: Statement
}

export function StatementBlock({ statement }: StatementBlockProps) {
  switch (statement.type) {
    case 'activityCall':
      return <ActivityCallBlock stmt={statement} />
    case 'workflowCall':
      return <WorkflowCallBlock stmt={statement} />
    case 'await':
      return <AwaitStmtBlock stmt={statement} />
    case 'awaitAll':
      return <AwaitAllBlockComponent stmt={statement} />
    case 'awaitOne':
      return <AwaitOneBlockComponent stmt={statement} />
    case 'switch':
      return <SwitchBlockComponent stmt={statement} />
    case 'if':
      return <IfBlock stmt={statement} />
    case 'for':
      return <ForBlock stmt={statement} />
    case 'return':
      return <ReturnBlock stmt={statement} />
    case 'close':
      return <CloseBlock stmt={statement} />
    case 'raw':
      return <RawBlock stmt={statement} />
    case 'break':
      return <SimpleBlock keyword="break" className="block-break" />
    case 'continue':
      return <SimpleBlock keyword="continue" className="block-continue" />
    case 'promise':
      return <PromiseBlock stmt={statement} />
    case 'set':
      return <SetBlock stmt={statement} />
    case 'unset':
      return <UnsetBlock stmt={statement} />
    case 'comment':
      return null // Skip comments in visualization
    default:
      return null
  }
}
