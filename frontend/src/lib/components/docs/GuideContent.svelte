<script lang="ts">
	import CodeBlock from './CodeBlock.svelte';

	export let pageId: string;

	const CURL_HEALTH    = `curl https://api.poi.trippier.dev/health`;
	const CURL_AUTH      = `curl "https://api.poi.trippier.dev/pois/search?lat=45.76&lng=4.83&radius=1500" \\\n  -H "X-API-Key: YOUR_API_KEY"`;
	const HTTP_AUTH      = `X-API-Key: YOUR_API_KEY`;
	const JSON_ERROR     = `{\n  "error": "invalid query: lat and lng are required for mode=radius"\n}`;
	const HTTP_RATELIMIT = `X-RateLimit-Limit:     60\nX-RateLimit-Remaining: 47\nX-RateLimit-Reset:     1716381234\nRetry-After:           12`;
	const CURL_BYOK      = `curl "https://api.poi.trippier.dev/pois/events?lat=45.76&lng=4.83" \\\n  -H "X-API-Key: YOUR_API_KEY" \\\n  -H "X-Ticketmaster-Key: YOUR_TM_KEY" \\\n  -H "X-Eventbrite-Token: YOUR_EB_TOKEN"`;
</script>

{#if pageId === 'quickstart'}
	<h1>Quickstart</h1>
	<p class="d-lead">Faites votre premier appel à l'API trippier en moins de deux minutes.</p>
	<h2>1. Récupérez l'API</h2>
	<p>Deux options, même schéma de routes :</p>
	<ul class="d-checklist">
		<li><strong>Cloud — </strong>créez un compte, récupérez votre clé API depuis le dashboard, 1 000 tokens offerts.</li>
		<li><strong>Self-hosted — </strong>une image Docker, vos serveurs, démarrez avec <code>POI_AUTH_DISABLED=true</code> pour désactiver l'auth.</li>
	</ul>
	<h2>2. Premier appel (sans auth)</h2>
	<p><code>/health</code> et <code>/pois/providers</code> ne nécessitent aucun token ni authentification :</p>
	<CodeBlock lang="bash" code={CURL_HEALTH} />
	<h2>3. Appel authentifié</h2>
	<p>Pour les routes facturées, passez votre clé via l'en-tête <code>X-API-Key</code> :</p>
	<CodeBlock lang="bash" code={CURL_AUTH} />
	<div class="d-callout">
		<strong>Routes gratuites.</strong>
		<span><code>GET /health</code> et <code>GET /pois/providers</code> sont publiques et ne consomment aucun token.</span>
	</div>

{:else if pageId === 'auth'}
	<h1>Authentification</h1>
	<p class="d-lead">L'API utilise des <strong>clés API</strong> transmises via l'en-tête <code>X-API-Key</code>. Générez-en plusieurs depuis le dashboard pour cloisonner vos environnements.</p>
	<h2>Format d'en-tête</h2>
	<CodeBlock lang="http" code={HTTP_AUTH} />
	<h2>Self-hosted sans auth</h2>
	<p>En self-hosting, définissez <code>POI_AUTH_DISABLED=true</code> pour désactiver entièrement la vérification des clés. Toutes les routes deviennent accessibles sans en-tête.</p>
	<h2>Rotation</h2>
	<p>Une clé peut être révoquée à tout moment depuis le dashboard. La révocation est immédiate.</p>
	<div class="d-callout warn">
		<strong>Ne committez jamais une clé.</strong>
		<span>Utilisez une variable d'environnement (<code>TRIPPIER_API_KEY</code>) ou un gestionnaire de secrets.</span>
	</div>

{:else if pageId === 'errors'}
	<h1>Erreurs</h1>
	<p class="d-lead">Les erreurs renvoient un JSON avec un champ <code>error</code> (string) et un code HTTP standard.</p>
	<h2>Format</h2>
	<CodeBlock lang="json" code={JSON_ERROR} />
	<h2>Codes courants</h2>
	<table class="d-table">
		<thead><tr><th>HTTP</th><th>Situation</th></tr></thead>
		<tbody>
			<tr><td>400</td><td>Paramètre absent, invalide ou combinaison incohérente (ex : <code>types</code> et <code>weights</code> simultanés).</td></tr>
			<tr><td>401</td><td>En-tête <code>X-API-Key</code> absent ou clé inconnue / révoquée.</td></tr>
			<tr><td>402</td><td>Solde de tokens insuffisant pour exécuter la requête.</td></tr>
			<tr><td>429</td><td>Rate-limit dépassé — attendez la fenêtre suivante.</td></tr>
			<tr><td>500</td><td>Erreur interne. Réessayez ; si persistant, ouvrez une issue.</td></tr>
		</tbody>
	</table>

{:else if pageId === 'ratelimit'}
	<h1>Rate-limit</h1>
	<p class="d-lead">Le rate-limit est basé sur la <strong>consommation de tokens</strong>, pas sur le nombre de requêtes. Chaque appel débite les tokens correspondants ; quand le solde est épuisé, l'API retourne un 402.</p>
	<h2>En-têtes renvoyés</h2>
	<CodeBlock lang="http" code={HTTP_RATELIMIT} />
	<h2>Self-hosted</h2>
	<p>En self-hosting avec <code>POI_AUTH_DISABLED=true</code>, aucun rate-limit n'est appliqué. Vous gérez votre propre quota côté infrastructure.</p>

{:else if pageId === 'tokens'}
	<h1>Tokens &amp; facturation</h1>
	<p class="d-lead">Les tokens sont décomptés à chaque appel facturé. 1 000 tokens sont offerts à l'inscription. En self-hosting, aucune notion de tokens.</p>
	<h2>Coût par route</h2>
	<table class="d-table">
		<thead><tr><th>Route</th><th>Tokens</th></tr></thead>
		<tbody>
			<tr><td><code>GET /pois/search</code> · <code>GET /pois/search/slim</code></td><td>1</td></tr>
			<tr><td><code>GET /pois/events</code> · <code>GET /pois/events/slim</code></td><td>10</td></tr>
			<tr><td><code>GET /pois/providers</code> · <code>GET /health</code></td><td>0 (gratuit)</td></tr>
		</tbody>
	</table>
	<div class="d-callout">
		<strong>Pourquoi les évènements coûtent 10 tokens ?</strong>
		<span>L'appel fan-out vers Ticketmaster et Eventbrite consomme des quotas API payants côté infrastructure. Le coût reflète cela.</span>
	</div>
	<h2>BYOK (Bring Your Own Key)</h2>
	<p>Apportez vos propres clés Ticketmaster / Eventbrite pour débloquer les évènements même sans solde suffisant :</p>
	<CodeBlock lang="bash" code={CURL_BYOK} />
{/if}

<style>
	h1 {
		font-size: 38px;
		font-weight: 600;
		letter-spacing: -0.025em;
		line-height: 1.1;
		margin: 0 0 16px;
	}
	h2 {
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.015em;
		margin: 48px 0 14px;
		padding-top: 16px;
		border-top: 1px solid var(--border);
	}
	p {
		font-size: 15px;
		line-height: 1.65;
		color: var(--text-2);
		margin: 0 0 14px;
	}
	:global(p code), :global(li code), :global(td code) {
		font-family: var(--font-mono);
		font-size: 12.5px;
		padding: 1px 5px;
		border-radius: 4px;
		background: color-mix(in oklch, var(--accent) 5%, var(--surface));
		border: 1px solid var(--border);
		color: var(--accent);
	}
	strong { color: var(--text); font-weight: 600; }
	.d-lead { font-size: 17px !important; color: var(--text-2); line-height: 1.6; margin-bottom: 24px !important; }
	.d-checklist { list-style: none; padding: 0; margin: 0 0 14px; }
	.d-checklist li {
		position: relative;
		padding: 6px 0 6px 22px;
		font-size: 14.5px;
		color: var(--text-2);
		line-height: 1.55;
	}
	.d-checklist li::before {
		content: '';
		position: absolute;
		left: 4px; top: 14px;
		width: 6px; height: 6px;
		border-radius: 50%;
		background: var(--accent);
	}
	.d-table { width: 100%; border-collapse: collapse; margin: 0 0 14px; }
	.d-table th,
	.d-table td {
		text-align: left;
		padding: 10px 12px;
		border-bottom: 1px solid var(--border);
		font-size: 13.5px;
		color: var(--text-2);
		vertical-align: top;
	}
	.d-table th {
		font-family: var(--font-mono);
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--text-3);
		font-weight: 500;
		background: color-mix(in oklch, var(--surface) 40%, transparent);
	}
	.d-table th:first-child,
	.d-table td:first-child { padding-left: 16px; }
	.d-callout {
		display: flex;
		gap: 12px;
		padding: 14px 18px;
		margin: 18px 0;
		border: 1px solid color-mix(in oklch, var(--accent) 25%, var(--border));
		border-left: 3px solid var(--accent);
		border-radius: var(--r-md);
		background: color-mix(in oklch, var(--accent) 5%, transparent);
		font-size: 13.5px;
		line-height: 1.55;
		color: var(--text-2);
	}
	.d-callout strong { margin-right: 6px; }
	.d-callout.warn {
		border-color: color-mix(in oklch, oklch(82% 0.12 80) 30%, var(--border));
		border-left-color: oklch(82% 0.12 80);
		background: color-mix(in oklch, oklch(82% 0.12 80) 4%, transparent);
	}
</style>
