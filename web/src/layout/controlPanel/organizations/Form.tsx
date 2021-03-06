import classnames from 'classnames';
import every from 'lodash/every';
import isNull from 'lodash/isNull';
import isUndefined from 'lodash/isUndefined';
import React, { useContext, useRef, useState } from 'react';

import { API } from '../../../api';
import { AppCtx, updateOrg } from '../../../context/AppCtx';
import { Organization, RefInputField, ResourceKind } from '../../../types';
import InputField from '../../common/InputField';
import InputFileField from '../../common/InputFileField';

interface FormValidation {
  isValid: boolean;
  organization: Organization | null;
}

interface Props {
  organization?: Organization;
  onSuccess?: () => void;
  onAuthError: () => void;
  setIsSending: (status: boolean) => void;
  setApiError?: React.Dispatch<React.SetStateAction<null>>;
}

const OrganizationForm = React.forwardRef<HTMLFormElement, Props>((props, ref) => {
  const { ctx, dispatch } = useContext(AppCtx);
  const [imageId, setImageId] = useState<string | undefined>(
    !isUndefined(props.organization) ? props.organization.logoImageId : undefined
  );
  const nameInput = useRef<RefInputField>(null);
  const homeUrlInput = useRef<RefInputField>(null);
  const [isValidated, setIsValidated] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);

  // Clean API error when form is focused after validation
  const cleanApiError = () => {
    if (!isNull(apiError)) {
      setApiError(null);
      if (!isUndefined(props.setApiError)) {
        props.setApiError(null);
      }
    }
  };

  async function handleOrganization(organization: Organization) {
    try {
      if (isUndefined(props.organization)) {
        await API.addOrganization(organization);
      } else {
        if (
          !isUndefined(ctx.prefs.controlPanel.selectedOrg) &&
          ctx.prefs.controlPanel.selectedOrg === organization.name
        ) {
          dispatch(updateOrg(organization.name));
        }
        await API.updateOrganization(organization);
      }
      props.setIsSending(false);
      if (!isUndefined(props.onSuccess)) {
        props.onSuccess();
      }
    } catch (err) {
      props.setIsSending(false);
      if (err.statusText !== 'ErrLoginRedirect') {
        let error = `An error occurred ${isUndefined(props.organization) ? 'adding' : 'updating'} the organization`;
        switch (err.status) {
          case 400:
            error += `: ${err.statusText}`;
            break;
          default:
            error += ', please try again later';
        }
        setApiError(error);
        if (!isUndefined(setApiError)) {
          setApiError(error);
        }
      } else {
        props.onAuthError();
      }
    }
  }

  const submitForm = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    e.stopPropagation();

    cleanApiError();
    props.setIsSending(true);
    if (e.currentTarget) {
      validateForm(e.currentTarget).then((validation: FormValidation) => {
        if (validation.isValid && !isNull(validation.organization)) {
          handleOrganization(validation.organization);
        } else {
          props.setIsSending(false);
        }
      });
    }
  };

  const validateForm = async (form: HTMLFormElement): Promise<FormValidation> => {
    let organization: Organization | null = null;

    return validateAllFields().then((isValid: boolean) => {
      if (isValid) {
        const formData = new FormData(form);
        organization = {
          name: formData.get('name') as string,
          displayName: formData.get('displayName') as string,
          homeUrl: formData.get('homeUrl') as string,
          description: formData.get('description') as string,
        };

        if (!isUndefined(imageId)) {
          organization.logoImageId = imageId;
        }
      }
      setIsValidated(true);
      return { isValid, organization };
    });
  };

  const validateAllFields = async (): Promise<boolean> => {
    return Promise.all([nameInput.current!.checkIsValid(), homeUrlInput.current!.checkIsValid()]).then(
      (res: boolean[]) => {
        return every(res, (isValid: boolean) => isValid);
      }
    );
  };

  return (
    <form
      ref={ref}
      data-testid="organizationForm"
      className={classnames('w-100', { 'needs-validation': !isValidated }, { 'was-validated': isValidated })}
      onFocus={cleanApiError}
      autoComplete="on"
      onSubmit={(e: React.FormEvent<HTMLFormElement>) => submitForm(e)}
      noValidate
    >
      <InputFileField
        name="logo"
        label="Logo"
        labelLegend={<small className="ml-1 font-italic">(Click on the image to update)</small>}
        value={imageId}
        onImageChange={(imageId: string) => setImageId(imageId)}
        onAuthError={props.onAuthError}
      />

      <InputField
        ref={nameInput}
        type="text"
        label="Name"
        labelLegend={<small className="ml-1 font-italic">(Required)</small>}
        name="name"
        value={!isUndefined(props.organization) ? props.organization.name : ''}
        readOnly={!isUndefined(props.organization)}
        invalidText={{
          default: 'This field is required',
          patternMismatch: 'Only lower case letters, numbers or hyphens',
          customError: 'There is another organization with this name',
        }}
        validateOnBlur
        checkAvailability={{
          isAvailable: true,
          resourceKind: ResourceKind.organizationName,
          excluded: !isUndefined(props.organization) ? [props.organization.name] : [],
        }}
        pattern="[a-z0-9-]+"
        autoComplete="off"
        required
      />

      <InputField
        type="text"
        label="Display name"
        name="displayName"
        value={
          !isUndefined(props.organization) && !isNull(props.organization.displayName)
            ? props.organization.displayName
            : ''
        }
      />

      <InputField
        ref={homeUrlInput}
        type="url"
        label="Home URL"
        name="homeUrl"
        invalidText={{
          default: 'Please enter a valid url',
        }}
        validateOnBlur
        value={
          !isUndefined(props.organization) && !isNull(props.organization.homeUrl) ? props.organization.homeUrl : ''
        }
      />

      <div className="form-group">
        <label htmlFor="description">Description</label>
        <textarea
          data-testid="descriptionTextarea"
          className="form-control"
          id="description"
          name="description"
          defaultValue={
            !isUndefined(props.organization) && !isNull(props.organization.description)
              ? props.organization.description
              : ''
          }
        />
      </div>

      {!isNull(apiError) && isUndefined(props.setApiError) && (
        <div className="alert alert-danger mt-3" role="alert">
          {apiError}
        </div>
      )}
    </form>
  );
});

export default OrganizationForm;
