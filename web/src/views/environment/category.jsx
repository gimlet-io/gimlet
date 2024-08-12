import { ArrowTopRightOnSquareIcon } from '@heroicons/react/24/solid';
import { RegistryWidget } from './registryWidget';
import HelmUI from "helm-react-ui";
import { EncryptedWidget } from './encryptedWidget';

export default function Category(props) {
  const { gimletClient, store } = props;
  const { environment } = props;
  const { stackConfig, stackDefinition, category, setValues, validationCallback } = props;

  if (!stackConfig) {
    return
  }

  const components = stackDefinition.components.filter(c => c.category === category)

  return (
    <div className="w-full space-y-8">
      {components.map(component =>
        <InfraComponent
          componentDefinition={component}
          config={stackConfig[component.variable] ?? {}}
          setValues={setValues}
          validationCallback={validationCallback}
          gimletClient={gimletClient}
          store={store}
          environment={environment}
        />
      )}
    </div>
  )
}

function InfraComponent(props) {
  const { componentDefinition, config, setValues, validationCallback } = props
  const { gimletClient, store } = props;
  const { environment } = props;

  const schema = typeof componentDefinition.schema === 'object'
    ? componentDefinition.schema
    : JSON.parse(componentDefinition.schema)
  const uiSchema = typeof componentDefinition.uiSchema === 'object'
    ? componentDefinition.uiSchema
    : JSON.parse(componentDefinition.uiSchema)

  const setVariableValues =
    (values, nonDefaultValues) => setValues(componentDefinition.variable, values, nonDefaultValues)

  const customFields = {
    "registryWidget": (props) => <RegistryWidget
      {...props}
      gimletClient={gimletClient}
      store={store}
      env={environment.name}
    />,
    "encryptedWidget": (props) => <EncryptedWidget
      {...props}
      gimletClient={gimletClient}
      store={store}
      env={environment.name}
    />,
  }

  return (
    <div>
      <div className='w-full card p-6 pb-8'>
        <HelmUI
          schema={schema}
          config={uiSchema}
          fields={customFields}
          values={config}
          setValues={setVariableValues}
          validate={true}
          validationCallback={(errors) => validationCallback(componentDefinition.variable, errors)}
        />
      </div>
      {uiSchema[0].metaData?.link && 
      <div className='-mt-2 learnMoreBox'>
        Learn more about <a href={uiSchema[0].metaData.link.href} className='learnMoreLink'>{uiSchema[0].metaData.link.label} <ArrowTopRightOnSquareIcon className="externalLinkIcon" aria-hidden="true" /></a>
      </div>
      }
    </div>
  )
}
