import { THEME, ThemeIcon, HANDLER_CONFIG, CLOSE_REASON_THEME, AWAIT_TARGET_THEME, WORKER_REF_THEME } from '../theme/temporal-theme'
import './style-guide.css'
import './blocks/blocks.css'

export function StyleGuide({ onClose }: { onClose: () => void }) {
  // Collect unique cssVarPrefix values for palette
  const prefixes = Array.from(new Set(
    (Object.values(THEME) as { cssVarPrefix: string }[]).map(t => t.cssVarPrefix)
  ))

  return (
    <div className="workflow-canvas">
      <div className="style-guide">
        <button className="style-guide-close" onClick={onClose}>
          Close (Ctrl+Shift+G)
        </button>
        <h1>TWF Style Guide</h1>

        {/* === Definitions === */}
        <div className="style-guide-category">
          <h2>Definitions</h2>

          {/* Namespace */}
          <div className="block block-namespace-def collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon block-icon-namespace">{THEME.namespace.icon}</span>
              <span className="block-keyword">namespace</span>
              <span className="block-signature">production (3 entries)</span>
            </div>
          </div>

          {/* Worker */}
          <div className="block block-worker-def collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon">{THEME.worker.icon}</span>
              <span className="block-keyword">worker</span>
              <span className="block-signature">main-worker (5 types)</span>
            </div>
          </div>

          {/* Worker refs */}
          {(['workflow', 'activity', 'service'] as const).map(refType => (
            <div key={refType} className={`worker-ref worker-ref-${refType} collapsed`}>
              <div className="worker-ref-header">
                <span className="block-toggle-placeholder" />
                <span className={`block-icon ${refType === 'service' ? 'block-icon-nexus-service' : ''}`}>
                  {WORKER_REF_THEME[refType].icon}
                </span>
                <span className="worker-ref-name">{WORKER_REF_THEME[refType].label}</span>
              </div>
            </div>
          ))}

          {/* Nexus Service */}
          <div className="block block-nexus-service-def collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon block-icon-nexus-service">{THEME.nexusService.icon}</span>
              <span className="block-keyword">service</span>
              <span className="block-signature">PaymentService (2 operations)</span>
            </div>
          </div>

          {/* Nexus Operations */}
          <div className="block block-nexus-operation nexus-operation-async collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon block-icon-nexus-operation">{THEME.nexusOperation.icon}</span>
              <span className="block-keyword">async</span>
              <span className="block-signature">ProcessPayment</span>
            </div>
          </div>
          <div className="block block-nexus-operation nexus-operation-sync collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon block-icon-nexus-operation">{THEME.nexusOperation.icon}</span>
              <span className="block-keyword">sync</span>
              <span className="block-signature">GetStatus(id string) &rarr; Status</span>
            </div>
          </div>

          {/* Workflow */}
          <div className="block block-workflow collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon"><ThemeIcon kind="workflow" /></span>
              <span className="block-keyword">workflow</span>
              <span className="block-signature">OrderWorkflow(input OrderInput) &rarr; OrderResult</span>
            </div>
          </div>

          {/* Activity */}
          <div className="block block-activity-def collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon"><ThemeIcon kind="activity" /></span>
              <span className="block-keyword">activity</span>
              <span className="block-signature">SendEmail(to string, body string)</span>
            </div>
          </div>
        </div>

        {/* === Calls & Awaits === */}
        <div className="style-guide-category">
          <h2>Calls & Awaits</h2>

          {/* Activity call */}
          <div className="block block-activity collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon"><ThemeIcon kind="activity" /></span>
              <span className="block-keyword">activity</span>
              <span className="block-signature">SendEmail(to, body) &rarr; err</span>
            </div>
          </div>

          {/* Workflow call */}
          <div className="block block-workflow-call collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon"><ThemeIcon kind="workflow" /></span>
              <span className="block-keyword">workflow</span>
              <span className="block-signature">ChildWorkflow(args) &rarr; result</span>
            </div>
          </div>

          {/* Workflow call detach */}
          <div className="block block-workflow-call block-mode-detach collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon"><ThemeIcon kind="workflow" /></span>
              <span className="block-keyword">detach workflow</span>
              <span className="block-signature">BackgroundJob(data)</span>
            </div>
          </div>

          {/* Nexus call */}
          <div className="block block-nexus-call collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon block-icon-nexus-call">{THEME.nexusCall.icon}</span>
              <span className="block-keyword">nexus</span>
              <span className="block-signature">ep PaymentService.Charge(amount) &rarr; receipt</span>
            </div>
          </div>

          {/* Await variants */}
          {(['timer', 'signal', 'update', 'activity', 'workflow', 'nexus', 'ident'] as const).map(kind => {
            const theme = AWAIT_TARGET_THEME[kind]
            const showIcon = kind !== 'activity' && kind !== 'workflow' && kind !== 'nexus'
            const sampleSig = kind === 'timer' ? '(30m)' :
              kind === 'signal' ? 'OrderCancelled' :
              kind === 'update' ? 'UpdateStatus(newStatus)' :
              kind === 'activity' ? 'ValidateInput(data)' :
              kind === 'workflow' ? 'SubWorkflow(args) → result' :
              kind === 'nexus' ? 'ep Svc.Op(args)' :
              'myPromise → result'
            return (
              <div key={kind} className={`block block-await-stmt block-await-stmt-${kind} collapsed`}>
                <div className="block-header">
                  <span className="block-toggle-placeholder" />
                  <span className="block-icon">{showIcon ? theme.icon : ''}</span>
                  <span className="block-keyword">await {kind === 'ident' ? '' : kind}</span>
                  <span className="block-signature">{sampleSig}</span>
                </div>
              </div>
            )
          })}

          {/* Await all */}
          <div className="block block-await-all collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon">{THEME.awaitAll.icon}</span>
              <span className="block-keyword">await all</span>
              <span className="block-signature">3 branch(es)</span>
            </div>
          </div>

          {/* Await one with option tags */}
          <div className="block block-await-one collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon-placeholder" />
              <span className="block-keyword">await one</span>
              <span className="block-signature">first of 2 cases</span>
            </div>
          </div>
          {/* Sample option tags */}
          <div className="tagged-composite">
            <div className="tagged-tag">
              <span className="tagged-tag-label">option</span>
            </div>
            <div className="tagged-content tagged-signal">
              <span className="block-toggle-placeholder" />
              <span className="tagged-icon">{THEME.signal.icon}</span>
              <span className="tagged-kind">signal</span>
              <span className="tagged-name">Cancel</span>
            </div>
          </div>
          <div className="tagged-composite">
            <div className="tagged-tag">
              <span className="tagged-tag-label">option</span>
            </div>
            <div className="tagged-content tagged-timer">
              <span className="block-toggle-placeholder" />
              <span className="tagged-icon">{THEME.timer.icon}</span>
              <span className="tagged-kind">timer</span>
              <span className="tagged-name">(1h)</span>
            </div>
          </div>
        </div>

        {/* === Handlers === */}
        <div className="style-guide-category">
          <h2>Handler Declarations</h2>
          {(Object.entries(HANDLER_CONFIG) as [string, { icon: string; keyword: string; cssClass: string }][]).map(([key, cfg]) => (
            <div key={key} className={`declaration ${cfg.cssClass}`}>
              <div className="declaration-header">
                <span className="block-toggle-placeholder" />
                <span className="declaration-icon">{cfg.icon}</span>
                <span className="declaration-keyword">{cfg.keyword}</span>
                <span className="declaration-name">Example{cfg.keyword.charAt(0).toUpperCase() + cfg.keyword.slice(1)}(params string)</span>
              </div>
            </div>
          ))}
        </div>

        {/* === Statements === */}
        <div className="style-guide-category">
          <h2>Statements</h2>

          {/* Return */}
          <div className="block block-return collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon">{THEME.return.icon}</span>
              <span className="block-keyword">return</span>
              <span className="block-signature">result</span>
            </div>
          </div>

          {/* Close variants */}
          {Object.entries(CLOSE_REASON_THEME).map(([reason, theme]) => {
            const statusClass = reason === 'continue_as_new' ? 'close-continue-as-new' : reason === 'fail' ? 'close-failed' : ''
            return (
              <div key={reason} className={`block block-close ${statusClass} collapsed`}>
                <div className="block-header">
                  <span className="block-toggle-placeholder" />
                  <span className="block-icon">{theme.icon}</span>
                  <span className="block-keyword">close</span>
                  <span className="block-signature">
                    <span className="close-reason">{reason}</span>
                  </span>
                </div>
              </div>
            )
          })}

          {/* Raw */}
          <div className="block block-raw collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon">{THEME.raw.icon}</span>
              <span className="block-code">logger.Info("processing order")</span>
            </div>
          </div>

          {/* Promise */}
          <div className="block block-promise collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon">{THEME.promise.icon}</span>
              <span className="block-keyword">promise</span>
              <span className="block-signature">emailFuture &larr; activity SendEmail(to, body)</span>
            </div>
          </div>

          {/* Set / Unset */}
          <div className="block block-set collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon">{THEME.conditionSet.icon}</span>
              <span className="block-keyword">set</span>
              <span className="block-signature">isReady</span>
            </div>
          </div>
          <div className="block block-unset collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon">{THEME.conditionUnset.icon}</span>
              <span className="block-keyword">unset</span>
              <span className="block-signature">isReady</span>
            </div>
          </div>

          {/* Break / Continue */}
          <div className="block block-break collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon">{THEME.breakContinue.icon}</span>
              <span className="block-keyword">break</span>
            </div>
          </div>
          <div className="block block-continue collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon">{THEME.breakContinue.icon}</span>
              <span className="block-keyword">continue</span>
            </div>
          </div>

          {/* For loop */}
          <div className="block block-for collapsed">
            <div className="block-header">
              <span className="block-toggle-placeholder" />
              <span className="block-icon">{THEME.forLoop.icon}</span>
              <span className="block-keyword">for</span>
              <span className="block-signature">item in items</span>
            </div>
          </div>
        </div>

        {/* === Color Palette === */}
        <div className="style-guide-category">
          <h2>Color Palette</h2>
          <div className="style-guide-palette">
            {prefixes.map(prefix => {
              const entry = Object.values(THEME).find(t => t.cssVarPrefix === prefix) as { label: string; cssVarPrefix: string; icon: string }
              return (
                <div
                  key={prefix}
                  className="style-guide-swatch"
                  style={{
                    background: `var(--block-${prefix}-bg)`,
                    border: `2px solid var(--block-${prefix}-border)`,
                    color: `var(--block-${prefix}-text)`,
                  }}
                >
                  <div className="style-guide-swatch-label">{entry.icon} {entry.label}</div>
                  <div className="style-guide-swatch-var">--block-{prefix}-*</div>
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}
