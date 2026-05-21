<svelte:head>
	<title>trippier/api · Connexion</title>
</svelte:head>

<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { auth } from '$lib/stores/auth';
	import { register, login, verifyCode, resendCode, getMe, ApiError } from '$lib/api';
	import { t, initLocale } from '$lib/i18n';
	import TopoBackground from '$lib/components/TopoBackground.svelte';
	import DocsNav from '$lib/components/docs/DocsNav.svelte';

	type Stage = 'form' | 'otp';
	type Mode  = 'login' | 'register';

	let mode: Mode   = 'login';
	let stage: Stage = 'form';
	let email        = '';
	let password     = '';
	let confirm      = '';
	let error        = '';
	let loading      = false;

	let digits: string[]          = Array(6).fill('');
	let digitEls: HTMLInputElement[] = [];

	let resendCooldown = 0;
	let resendOk = false;
	let resendTimer: ReturnType<typeof setInterval> | null = null;

	function startCooldown() {
		resendCooldown = 30;
		resendTimer = setInterval(() => {
			resendCooldown--;
			if (resendCooldown <= 0) {
				clearInterval(resendTimer!);
				resendTimer = null;
			}
		}, 1000);
	}

	async function handleResend() {
		if (resendCooldown > 0 || loading) return;
		resendOk = false;
		loading = true;
		try {
			await resendCode(email);
			resendOk = true;
			startCooldown();
		} catch (e) {
			error = e instanceof ApiError ? e.message : $t('login_error_generic');
		} finally {
			loading = false;
		}
	}

	onMount(async () => {
		initLocale();
		const stored = auth.getStoredToken();
		if (!stored) return;
		try {
			const user = await getMe(stored);
			auth.init(stored, user);
			goto('/dashboard');
		} catch {}
	});

	async function submitForm() {
		error = '';
		if (mode === 'register' && password !== confirm) {
			error = $t('login_no_match');
			return;
		}
		loading = true;
		try {
			if (mode === 'register') {
				await register(email, password);
				stage  = 'otp';
				digits = Array(6).fill('');
				setTimeout(() => digitEls[0]?.focus(), 50);
			} else {
				const token = await login(email, password);
				const user  = await getMe(token);
				auth.init(token, user);
				auth.storeToken(token);
				goto('/dashboard');
			}
		} catch (e) {
			error = e instanceof ApiError ? e.message : $t('login_error_generic');
		} finally {
			loading = false;
		}
	}

	async function submitOTP() {
		error = '';
		const code = digits.join('');
		if (code.length !== 6) { error = 'Entrez les 6 chiffres'; return; }
		loading = true;
		try {
			const token = await verifyCode(email, code);
			const user  = await getMe(token);
			auth.init(token, user);
			auth.storeToken(token);
			goto('/dashboard');
		} catch (e) {
			error  = e instanceof ApiError ? e.message : $t('login_invalid_code');
			digits = Array(6).fill('');
			setTimeout(() => digitEls[0]?.focus(), 50);
		} finally {
			loading = false;
		}
	}

	function handleDigitInput(i: number) {
		digits[i] = digits[i].replace(/\D/g, '').slice(-1);
		if (digits[i] && i < 5) digitEls[i + 1]?.focus();
	}

	function handleDigitKeydown(e: KeyboardEvent, i: number) {
		if (e.key === 'Backspace') {
			if (digits[i]) { digits[i] = ''; }
			else if (i > 0) { digits[i - 1] = ''; digitEls[i - 1]?.focus(); }
			e.preventDefault();
		}
		if (e.key === 'ArrowLeft'  && i > 0) digitEls[i - 1]?.focus();
		if (e.key === 'ArrowRight' && i < 5) digitEls[i + 1]?.focus();
	}

	function handleDigitPaste(e: ClipboardEvent) {
		e.preventDefault();
		const pasted = (e.clipboardData?.getData('text') ?? '').replace(/\D/g, '').slice(0, 6);
		digits = [...pasted.split(''), ...Array(6).fill('')].slice(0, 6);
		digitEls[Math.min(pasted.length, 5)]?.focus();
	}

	function switchMode(m: Mode) {
		mode    = m;
		error   = '';
		confirm = '';
		if (stage === 'otp') stage = 'form';
	}
</script>

<div class="lp">
	{#if browser}<TopoBackground density={28} opacity={0.24} color="#34d39c" />{/if}

	<DocsNav />

	<div class="lp-center">
		<div class="lp-card">
			{#if stage === 'form'}
				<div class="lp-tabs">
					<button class="lp-tab" class:active={mode === 'login'}    on:click={() => switchMode('login')}>{$t('login_tab_login')}</button>
					<button class="lp-tab" class:active={mode === 'register'} on:click={() => switchMode('register')}>{$t('login_tab_register')}</button>
				</div>

				<div class="lp-head">
					<h1>{mode === 'login' ? $t('login_title_login') : $t('login_title_register')}</h1>
					<p class="lp-sub">
						{mode === 'login' ? $t('login_sub_login') : $t('login_sub_register')}
					</p>
				</div>

				{#if error}<p class="lp-error">{error}</p>{/if}

				<form on:submit|preventDefault={submitForm} class="lp-form">
					<label class="lp-field">
						<span>{$t('login_email')}</span>
						<input type="email" bind:value={email} placeholder="you@example.com" required autocomplete="email" />
					</label>
					<label class="lp-field">
						<span>{$t('login_password')}</span>
						<input type="password" bind:value={password} placeholder="••••••••" required autocomplete={mode === 'login' ? 'current-password' : 'new-password'} />
					</label>
					{#if mode === 'register'}
						<label class="lp-field">
							<span>{$t('login_confirm')}</span>
							<input type="password" bind:value={confirm} placeholder="••••••••" required autocomplete="new-password" />
						</label>
					{/if}
					<button class="lp-submit" type="submit" disabled={loading}>
						{#if loading}…{:else if mode === 'register'}{$t('login_submit_register')}{:else}{$t('login_submit_login')}{/if}
					</button>
				</form>

			{:else}
				<button class="lp-back" on:click={() => { stage = 'form'; error = ''; }}>
					<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M19 12H5M12 5l-7 7 7 7"/></svg>
					{$t('login_otp_back')}
				</button>

				<div class="lp-head">
					<h1>{$t('login_otp_title')}</h1>
					<p class="lp-sub">{$t('login_otp_sub')} <strong>{email}</strong>. {$t('login_otp_expires')}</p>
				</div>

				{#if error}<p class="lp-error">{error}</p>{/if}

				<form on:submit|preventDefault={submitOTP} class="lp-form">
					<div class="lp-otp-wrap">
						<p class="lp-otp-label">{$t('login_otp_label')}</p>
						<div class="lp-otp">
							{#each digits as _, i}
								{#if i === 3}
									<span class="lp-otp-sep">·</span>
								{/if}
								<input
									type="text"
									inputmode="numeric"
									maxlength="2"
									class="lp-digit"
									bind:value={digits[i]}
									bind:this={digitEls[i]}
									on:input={() => handleDigitInput(i)}
									on:keydown={(e) => handleDigitKeydown(e, i)}
									on:paste={handleDigitPaste}
									autocomplete="one-time-code"
								/>
							{/each}
						</div>
					</div>
					<button class="lp-submit" type="submit" disabled={loading || digits.join('').length < 6}>
						{loading ? '…' : $t('login_otp_verify')}
					</button>
				</form>

				<p class="lp-hint">
					{#if resendOk}
						{$t('login_otp_resend_ok')}
					{:else}
						{$t('login_otp_hint')}
						<button class="lp-resend" on:click={handleResend} disabled={resendCooldown > 0 || loading}>
							{resendCooldown > 0 ? $t('login_otp_resend_wait').replace('{s}', String(resendCooldown)) : $t('login_otp_resend')}
						</button>
					{/if}
				</p>
			{/if}
		</div>
	</div>
</div>

<style>
	.lp {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		position: relative;
	}

	.lp-center {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 40px 16px 80px;
		position: relative;
		z-index: 1;
	}

	.lp-card {
		width: 100%;
		max-width: 400px;
		background: var(--bg-2);
		border: 1px solid var(--border);
		border-radius: var(--r-lg);
		padding: 32px;
		display: flex;
		flex-direction: column;
		gap: 20px;
	}

	.lp-tabs {
		display: flex;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 3px;
		gap: 2px;
	}
	.lp-tab {
		flex: 1;
		background: none;
		border: none;
		border-radius: calc(var(--r-md) - 2px);
		color: var(--text-3);
		font-family: var(--font-sans);
		font-size: 13px;
		font-weight: 500;
		padding: 7px;
		cursor: pointer;
		transition: background .12s ease, color .12s ease;
	}
	.lp-tab.active { background: var(--bg-2); color: var(--text); }

	.lp-head { display: flex; flex-direction: column; gap: 6px; }
	.lp-head h1 {
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.02em;
		line-height: 1.15;
	}
	.lp-sub { font-size: 13.5px; color: var(--text-2); line-height: 1.55; }
	.lp-sub strong { color: var(--text); font-weight: 500; }

	.lp-error {
		font-size: 13px;
		color: oklch(72% 0.16 25);
		background: oklch(72% 0.16 25 / 0.08);
		border: 1px solid oklch(72% 0.16 25 / 0.25);
		border-radius: var(--r-md);
		padding: 10px 14px;
		margin: 0;
	}

	.lp-form { display: flex; flex-direction: column; gap: 12px; }

	.lp-field {
		display: flex;
		flex-direction: column;
		gap: 5px;
	}
	.lp-field span {
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-3);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	.lp-field input {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--r-md);
		padding: 10px 12px;
		font-family: var(--font-sans);
		font-size: 14px;
		color: var(--text);
		outline: none;
		transition: border-color .12s ease;
	}
	.lp-field input:focus { border-color: var(--accent); }
	.lp-field input::placeholder { color: var(--text-3); }

	.lp-submit {
		width: 100%;
		padding: 11px;
		background: var(--accent);
		color: #08120e;
		border: none;
		border-radius: var(--r-md);
		font-family: var(--font-sans);
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
		transition: filter .15s ease;
		margin-top: 4px;
	}
	.lp-submit:hover:not(:disabled) { filter: brightness(1.08); }
	.lp-submit:disabled { opacity: 0.45; cursor: not-allowed; }

	.lp-back {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-family: var(--font-mono);
		font-size: 11.5px;
		color: var(--text-3);
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
		transition: color .12s ease;
	}
	.lp-back:hover { color: var(--text); }

	.lp-otp-wrap {
		background: var(--bg);
		border: 1px solid var(--border);
		border-left: 3px solid var(--accent);
		border-radius: var(--r-md);
		padding: 20px 16px 20px 14px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 14px;
	}
	.lp-otp-label {
		font-family: var(--font-mono);
		font-size: 10.5px;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--text-3);
		align-self: flex-start;
	}
	.lp-otp {
		display: flex;
		gap: 6px;
		align-items: center;
	}
	.lp-otp-sep {
		font-family: var(--font-mono);
		font-size: 20px;
		color: var(--text-3);
		width: 10px;
		text-align: center;
	}
	.lp-digit {
		width: 44px;
		height: 56px;
		flex-shrink: 0;
		text-align: center;
		font-size: 26px;
		font-weight: 600;
		font-family: var(--font-mono);
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: 8px;
		color: var(--text);
		outline: none;
		caret-color: transparent;
		transition: border-color .12s ease;
		padding: 0;
	}
	.lp-digit:focus { border-color: var(--accent); }

	.lp-hint {
		font-size: 12px;
		color: var(--text-3);
		text-align: center;
		line-height: 1.5;
	}

	.lp-resend {
		background: none;
		border: none;
		padding: 0;
		font-size: 12px;
		font-family: var(--font-sans);
		color: var(--accent);
		cursor: pointer;
		text-decoration: underline;
		text-underline-offset: 2px;
		transition: opacity .12s ease;
	}
	.lp-resend:disabled {
		color: var(--text-3);
		text-decoration: none;
		cursor: default;
	}
</style>
